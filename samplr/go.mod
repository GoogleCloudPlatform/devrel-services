module devrel/cloud/devrel-github-service/samplr

require (
	cloud.google.com/go v0.40.0
	devrel/cloud/devrel-github-service/drghs v0.0.0
	devrel/cloud/devrel-github-service/git-go v0.0.0
	devrel/cloud/devrel-github-service/repos v0.0.0
	devrel/cloud/devrel-github-service/rtr v0.0.0 // indirect
	devrel/cloud/devrel-github-service/sprvsr v0.0.0

	github.com/cespare/trie v0.0.0-20150610204604-3fe1a95cbba9 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/cel-go v0.3.0
	github.com/google/go-cmp v0.3.0
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/toqueteos/trie v0.0.0-20150530104557-56fed4a05683 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/grpc v1.23.0
	gopkg.in/src-d/enry.v1 v1.6.7
	gopkg.in/toqueteos/substring.v1 v1.0.2 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190528154508-67ef80593b24
	k8s.io/apimachinery v0.0.0-20190528154326-e59c2fb0a8e5
	k8s.io/client-go v0.0.0-20190528154735-79226fe1949a
)

replace devrel/cloud/devrel-github-service/drghs => ../drghs

replace devrel/cloud/devrel-github-service/git-go => ./git-go

replace devrel/cloud/devrel-github-service/repos => ../repos

replace devrel/cloud/devrel-github-service/rtr => ../rtr

replace devrel/cloud/devrel-github-service/sprvsr => ../sprvsr

go 1.13
