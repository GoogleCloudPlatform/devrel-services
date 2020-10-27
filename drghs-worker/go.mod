// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

module github.com/GoogleCloudPlatform/devrel-services/drghs-worker

require (
	cloud.google.com/go v0.61.0
	cloud.google.com/go/storage v1.10.0
	github.com/GoogleCloudPlatform/devrel-services/drghs v0.0.0-20200730153546-93a9c4fcaf2c
	github.com/GoogleCloudPlatform/devrel-services/repos v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/rtr v0.0.0 // indirect
	github.com/GoogleCloudPlatform/devrel-services/sprvsr v0.0.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.2.1
	github.com/aclements/go-gg v0.0.0-20170323211221-abd1f791f5ee // indirect
	github.com/aclements/go-moremath v0.0.0-20190830160640-d16893ddf098 // indirect
	github.com/ajstarks/deck v0.0.0-20200217041847-5bf3c34cfe40 // indirect
	github.com/ajstarks/svgo v0.0.0-20200204031535-0cbcf57ea1d8 // indirect
	github.com/antlr/antlr4 v0.0.0-20200712162734-eb1adaa8a7a6 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cznic/cc v0.0.0-20181122101902-d673e9b70d4d // indirect
	github.com/cznic/fileutil v0.0.0-20181122101858-4d67cfea8c87 // indirect
	github.com/cznic/golex v0.0.0-20181122101858-9c343928389c // indirect
	github.com/cznic/internal v0.0.0-20181122101858-3279554c546e // indirect
	github.com/cznic/ir v0.0.0-20181122101859-da7ba2ecce8b // indirect
	github.com/cznic/lex v0.0.0-20181122101858-ce0fb5e9bb1b // indirect
	github.com/cznic/lexer v0.0.0-20181122101858-e884d4bd112e // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/cznic/xc v0.0.0-20181122101856-45b06973881e // indirect
	github.com/dimchansky/utfbom v1.1.0 // indirect
	github.com/disintegration/gift v1.2.1 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/emicklei/go-restful v2.11.2+incompatible // indirect
	github.com/emicklei/go-restful-openapi v1.3.0 // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/go-gl/glfw v0.0.0-20200222043503-6f7a984d4dc4 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gonum/blas v0.0.0-20181208220705-f22b278b28ac // indirect
	github.com/gonum/floats v0.0.0-20181209220543-c233463c7e82 // indirect
	github.com/gonum/internal v0.0.0-20181124074243-f884aa714029 // indirect
	github.com/gonum/lapack v0.0.0-20181123203213-e4cdc5a0bff9 // indirect
	github.com/gonum/matrix v0.0.0-20181209220409-c518dec07be9 // indirect
	github.com/google/cel-go v0.5.1
	github.com/google/go-cmp v0.5.1
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/schema v1.1.0 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/graph-gophers/graphql-go v0.0.0-20200207002730-8334863f2c8b // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.1 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334 // indirect
	github.com/inconshreveable/go-vhost v0.0.0-20160627193104-06d84117953b // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/jung-kurt/gofpdf v1.16.2 // indirect
	github.com/lyft/protoc-gen-star v0.4.14 // indirect
	github.com/matryer/is v1.2.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d // indirect
	github.com/pkg/sftp v1.11.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20190728182440-6a916e37a237 // indirect
	github.com/rogpeppe/go-charset v0.0.0-20190617161244-0dc95cdf6f31 // indirect
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd // indirect
	github.com/shurcooL/githubv4 v0.0.0-20191102174205-af46314aec7b
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/syndtr/goleveldb v1.0.0 // indirect
	go.opentelemetry.io/otel v0.9.0
	go4.org v0.0.0-20200411211856-f5505b9728dd // indirect
	golang.org/x/build v0.0.0-20200730154719-cb64255c8b23
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/exp v0.0.0-20200513190911-00229845015e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200728102440-3e129f6d46b1 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.0.0-20200729194436-6467de6f59a7 // indirect
	google.golang.org/api v0.29.0
	google.golang.org/genproto v0.0.0-20200730144737-007c33dbd381
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/vmihailenco/msgpack.v2 v2.9.1 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	grpc.go4.org v0.0.0-20170609214715-11d0a25b4919
	k8s.io/api v0.0.0-20190528154508-67ef80593b24
	k8s.io/apimachinery v0.0.0-20190528154326-e59c2fb0a8e5
	k8s.io/client-go v0.0.0-20190528154735-79226fe1949a
	k8s.io/gengo v0.0.0-20200205140755-e0e292d8aa12 // indirect
	sigs.k8s.io/structured-merge-diff v1.0.2 // indirect
)

replace github.com/GoogleCloudPlatform/devrel-services/drghs => ../drghs

replace github.com/GoogleCloudPlatform/devrel-services/rtr => ../rtr

replace github.com/GoogleCloudPlatform/devrel-services/sprvsr => ../sprvsr

replace github.com/GoogleCloudPlatform/devrel-services/repos => ../repos

replace golang.org/x/build => github.com/orthros/build v0.0.0-20200730160535-a45e4470b022

go 1.13
