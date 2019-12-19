module github.com/GoogleCloudPlatform/devrel-services/devrelservices-admin

go 1.13

require (
	cloud.google.com/go v0.49.0
	github.com/GoogleCloudPlatform/devrel-services/drghs v0.0.0-20191204181555-5cde750c6624
	github.com/sirupsen/logrus v1.4.2
	google.golang.org/grpc v1.25.1
)

replace github.com/GoogleCloudPlatform/devrel-services/drghs => ../drghs

replace github.com/GoogleCloudPlatform/devrel-services/rtr => ../rtr

replace github.com/GoogleCloudPlatform/devrel-services/sprvsr => ../sprvsr

replace github.com/GoogleCloudPlatform/devrel-services/repos => ../repos
