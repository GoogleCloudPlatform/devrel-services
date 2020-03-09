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

module github.com/GoogleCloudPlatform/devrel-services/sprvsr

go 1.12

require (
	cloud.google.com/go v0.40.0
	github.com/GoogleCloudPlatform/devrel-services/repos v0.0.0

	github.com/deckarep/golang-set v1.7.1
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/mux v1.7.2
	github.com/gregjones/httpcache v0.0.0-20190203031600-7a902570cb17 // indirect
	github.com/matryer/is v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/negroni v1.0.0
	go4.org v0.0.0-20181109185143-00e24f1b2599 // indirect
	golang.org/x/build v0.0.0-20190201181641-63986c177d1f
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	grpc.go4.org v0.0.0-20170609214715-11d0a25b4919
	k8s.io/api v0.0.0-20190528154508-67ef80593b24
	k8s.io/apimachinery v0.0.0-20190528154326-e59c2fb0a8e5
	k8s.io/client-go v0.0.0-20190528154735-79226fe1949a
	k8s.io/utils v0.0.0-20190529001817-6999998975a7 // indirect
)

replace github.com/GoogleCloudPlatform/devrel-services/repos => ../repos
