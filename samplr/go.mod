module github.com/GoogleCloudPlatform/devrel-services/samplr

require (
	cloud.google.com/go v0.40.0
	github.com/GoogleCloudPlatform/devrel-services/drghs v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/git-go v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/repos v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/rtr v0.0.0 // indirect
	github.com/GoogleCloudPlatform/devrel-services/sprvsr v0.0.0

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

replace github.com/GoogleCloudPlatform/devrel-services/drghs => ../drghs

replace github.com/GoogleCloudPlatform/devrel-services/git-go => ./git-go

replace github.com/GoogleCloudPlatform/devrel-services/repos => ../repos

replace github.com/GoogleCloudPlatform/devrel-services/rtr => ../rtr

replace github.com/GoogleCloudPlatform/devrel-services/sprvsr => ../sprvsr

go 1.13
