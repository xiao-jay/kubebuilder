package cli

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/config/store/yaml"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	kustomizev2scaffolds "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v2/scaffolds"
	golangv4 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v4"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v4/scaffolds"
)

var (
	noBuckUp   bool
	inputDir   string
	outputDir  string
	backUpPath string
)

const (
	backupDirName = ".backup"
)

func (c *CLI) newAlphaGenerateCommand() *cobra.Command {
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Copy the project to the directory project-v3/<project-name> and re-generate the project based on the PROJECT file config",
		Long: `This command is a helper for you upgrade your project to the latest versions scaffold.
		It will:
			- Create a new directory named project-v3/<project-name>
			- Then, will remove all content under the project directory
			- Re-generate the whole project based on the Project file data
		Therefore, you can use it to upgrade your project since as a follow up you would need to 
		only compare the project copied to project-v3/<project-name> in order to add on top again all
		your code implementation and customizations.`,
		PreRunE: c.validation,
		RunE:    c.runE,
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	generate.Flags().BoolVarP(&noBuckUp, "no-backup", "", false, "re-Scaffold will not backup your project file if true")
	generate.Flags().StringVarP(&inputDir, "input-dir", "i", dir, "path where the PROJECT file can be found")
	generate.Flags().StringVarP(&outputDir, "output-dir", "o", dir, "path where the project should be re-scaffold")
	generate.Flags().StringVarP(&backUpPath, "backup-path", "b", dir, "path-where the current version of the project should be copied as backup")
	return generate
}

func (c *CLI) validation(cmd *cobra.Command, args []string) error {
	fmt.Printf("inputdir:%s,outpuddif:%s,backupdir:%s\n", inputDir, outputDir, backUpPath)
	// check existed PROJECT file
	fmt.Println("Check if PROJECT exists")
	PROJECTPath := filepath.Join(inputDir, "PROJECT")
	if _, err := c.fs.FS.Stat(PROJECTPath); err != nil {
		return err
	}

	//check if PROJECT file is valid
	if _, err := readPROJECT(inputDir); err != nil {
		return err
	}

	//checkout version of plugin is same as project version, is same will not next step
	fmt.Println("check plugin,resolvedPlugins:", c.resolvedPlugins)
	if len(c.resolvedPlugins) == 0 {
		return fmt.Errorf("no plugin found")
	}

	for _, v := range c.resolvedPlugins {
		if !(v.Name() == "go.kubebuilder.io" && v.Version().String() == "v4") {
			return fmt.Errorf("now only support go.kubebuilder.io/v4")
		}
	}
	return nil
}

func (c *CLI) runE(cmd *cobra.Command, args []string) error {
	fmt.Println("enter run")
	configInfo, err := readPROJECT(inputDir)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("config:", configInfo)

	if err = backupProject(inputDir, backUpPath); err != nil {
		return err
	}
	fs := afero.NewOsFs()
	c.fs = machinery.Filesystem{FS: fs}
	goPluins := golangv4.Plugin{}

	err = kubebuilderInit(configInfo, c.plugins)
	if err != nil {
		return err
	}

	err = c.reCreateApi(fs, configInfo, goPluins)
	if err != nil {
		return err
	}
	//err = c.reCreateWebhook(fs, configInfo, goPluins)
	//if err != nil {
	//	return err
	//}

	return nil
}

func kubebuilderInit(config config.Config, plugins map[string]plugin.Plugin) error {
	var args []string
	args = append(args, "init")
	args = append(args, generateInitArgs(config)...)
	if err := os.Chdir(outputDir); err != nil {
		return nil
	}
	return util.RunCmd("kubebuilder init", "kubebuilder", args...)
}

func generateInitArgs(config config.Config) []string {
	var args []string
	args = append(args, "--plugins", "go.kubebuilder.io/v4")

	domain := config.GetDomain()
	if len(domain) > 0 {
		args = append(args, "--domain")
		args = append(args, domain)
	}

	repo := config.GetRepository()
	if len(repo) > 0 {
		args = append(args, "--repo")
		args = append(args, repo)
	}
	return args
}

func backupProject(inputDir, backupPath string) error {
	if !noBuckUp {
		output := filepath.Join(backupPath, "../", backupDirName)
		if err := os.MkdirAll(output, os.ModePerm); err != nil {
			return err
		}
		if err := util.RunCmd(fmt.Sprintf("Copying all files to %s", output), "cp", "-a", inputDir, output); err != nil {
			return err
		}
	}

	dir, _ := os.ReadDir(inputDir)
	for _, d := range dir {
		if d.Name() == "." || d.Name() == ".." {
			continue
		}
		os.RemoveAll(path.Join([]string{inputDir, d.Name()}...))
	}
	return nil
}

func readPROJECT(path string) (config.Config, error) {
	store := yaml.New(machinery.Filesystem{FS: afero.NewOsFs()})
	if err := store.LoadFrom(path + "/PROJECT"); err != nil {
		return nil, err
	}
	return store.Config(), nil
}

//	func (c *CLI) reInitProject(fs afero.Fs, configInfo config.Config, plugin golangv4.Plugin) error {
//		logrus.Info("reInitProject")
//
//		scaffold := scaffolds.NewInitScaffolder(configInfo, "apache2", "")
//		scaffold.InjectFS(machinery.Filesystem{FS: fs})
//		return scaffold.Scaffold()
//	}

func (c *CLI) reCreateApi(fs afero.Fs, configInfo config.Config, pluginv4 plugin.Plugin) error {
	logrus.Info("reCreate APi")
	resources, err := configInfo.GetResources()
	if err != nil {
		return err
	}
	for _, resource := range resources {
		scaffold := scaffolds.NewAPIScaffolder(configInfo, resource, false)
		scaffold.InjectFS(machinery.Filesystem{FS: fs})
		err := scaffold.Scaffold()
		if err != nil {
			return err
		}
		kscaffold := kustomizev2scaffolds.NewAPIScaffolder(configInfo, resource, false)
		kscaffold.InjectFS(machinery.Filesystem{FS: fs})
		if err := kscaffold.Scaffold(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) reCreateWebhook(fs afero.Fs, configInfo config.Config, plugin golangv4.Plugin) error {
	logrus.Info("reCreate Webhook")
	resources, err := configInfo.GetResources()
	if err != nil {
		return err
	}

	for _, resource := range resources {
		fmt.Printf("resource:%+v", resource)
		if resource.Webhooks == nil {
			fmt.Println("resource not have webhook")
			return nil
		}
		scaffold := scaffolds.NewWebhookScaffolder(configInfo, resource, true)
		scaffold.InjectFS(machinery.Filesystem{FS: fs})
		err := scaffold.Scaffold()
		if err != nil {
			return err
		}

		kscaffold := kustomizev2scaffolds.NewWebhookScaffolder(configInfo, resource, true)
		kscaffold.InjectFS(machinery.Filesystem{FS: fs})
		if err := kscaffold.Scaffold(); err != nil {
			return err
		}
	}
	return nil
}
