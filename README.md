## kubebuilder alpha generate cmd demo

This example demonstrates how to use kubebuilder alpha generate cmd to re-scaffold a newest project

usage:
```shell
git clone -b re-scaffold https://github.com/xiao-jay/kubebuilder.git
make build # make build will run mv bin/kubebuilder $GOPATH/bin 
kubebuilder alpha generate -i testdata/project-v3 -o testdata/project-v3 -b testdata/project-v3 --plugins go.kubebuilder.io/v4
cd testdata/project-v3
make install
```

make install result:
```shell
➜  project-v3 git:(re-scaffold) ✗ make install
mkdir -p /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin
test -s /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/controller-gen && /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/controller-gen --version | grep -q v0.11.3 || \
        GOBIN=/Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.3
/Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
test -s /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/kustomize || { curl -Ss "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" --output install_kustomize.sh && bash install_kustomize.sh 5.0.0 /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin; rm install_kustomize.sh; }
v5.0.0
kustomize installed to /Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/kustomize
/Users/jie/Desktop/golang/kubebuilder/testdata/project-v3/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/admirales.crew.testproject.org created
customresourcedefinition.apiextensions.k8s.io/captains.crew.testproject.org created
customresourcedefinition.apiextensions.k8s.io/firstmates.crew.testproject.org created
```

Describe
```shell
➜  kubebuilder git:(re-scaffold) ✗ kubebuilder alpha generate -h
This command is a helper for you upgrade your project to the latest versions scaffold.
                It will:
                        - Create a new directory named project-v3/<project-name>
                        - Then, will remove all content under the project directory
                        - Re-generate the whole project based on the Project file data
                Therefore, you can use it to upgrade your project since as a follow up you would need to 
                only compare the project copied to project-v3/<project-name> in order to add on top again all
                your code implementation and customizations.

Usage:
  kubebuilder alpha generate [flags]

Flags:
  -b, --backup-path string   path-where the current version of the project should be copied as backup (default "/Users/jie/Desktop/golang/kubebuilder")
  -h, --help                 help for generate
  -i, --input-dir string     path where the PROJECT file can be found (default "/Users/jie/Desktop/golang/kubebuilder")
      --no-backup            re-Scaffold will not backup your project file if true
  -o, --output-dir string    path where the project should be re-scaffold (default "/Users/jie/Desktop/golang/kubebuilder")

Global Flags:
```
