module github.com/GoogleCloudPlatform/devrel-services/leif

go 1.14

require (
	github.com/GoogleCloudPlatform/devrel-services/drghs v0.0.0-20200723024905-6c479f56d135
	github.com/GoogleCloudPlatform/devrel-services/repos v0.0.0-20200720163603-c134bef7ad58
	github.com/golang/protobuf v1.4.2
	github.com/google/cel-go v0.5.1
	github.com/google/go-cmp v0.5.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v32 v32.0.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/mitchellh/mapstructure v1.3.2
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.25.0
)

replace github.com/GoogleCloudPlatform/devrel-services/drghs => ../drghs
