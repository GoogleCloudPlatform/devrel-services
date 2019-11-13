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
	cloud.google.com/go v0.40.0
	github.com/GoogleCloudPlatform/devrel-services/drghs v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/devrel-services/repos v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/rtr v0.0.0
	github.com/GoogleCloudPlatform/devrel-services/sprvsr v0.0.0
	github.com/golang/protobuf v1.3.2
	github.com/google/cel-go v0.3.0
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/mux v1.7.2
	github.com/matryer/is v1.2.0
	github.com/shurcooL/githubv4 v0.0.0-20191102174205-af46314aec7b
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/negroni v1.0.0
	golang.org/x/build v0.0.0-20190201181641-63986c177d1f
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/api v0.6.0
	google.golang.org/grpc v1.24.0
	grpc.go4.org v0.0.0-20170609214715-11d0a25b4919
	k8s.io/api v0.0.0-20190528154508-67ef80593b24
	k8s.io/apimachinery v0.0.0-20190528154326-e59c2fb0a8e5
	k8s.io/client-go v0.0.0-20190528154735-79226fe1949a
)

replace github.com/GoogleCloudPlatform/devrel-services/drghs => ../drghs

replace github.com/GoogleCloudPlatform/devrel-services/rtr => ../rtr

replace github.com/GoogleCloudPlatform/devrel-services/sprvsr => ../sprvsr

replace github.com/GoogleCloudPlatform/devrel-services/repos => ../repos

go 1.13
