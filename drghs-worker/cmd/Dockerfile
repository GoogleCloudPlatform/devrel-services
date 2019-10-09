# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.13 as build

ENV GO111MODULE=on

WORKDIR /src/devrel/cloud/devrel-github-service
COPY . .

# BEGIN Jank to get an upstream patch into the build
WORKDIR /src/devrel/cloud/devrel-github-service
RUN git clone https://go.googlesource.com/build build
WORKDIR /src/devrel/cloud/devrel-github-service/build
RUN git fetch https://go.googlesource.com/build refs/changes/38/176638/5 && git checkout FETCH_HEAD

RUN echo "replace golang.org/x/build => ../build " >> ../drghs-worker/go.mod

WORKDIR /src/devrel/cloud/devrel-github-service/drghs-worker

# END Jank

RUN go install devrel/cloud/devrel-github-service/drghs-worker/cmd

FROM golang:1.11
COPY --from=build /go/bin/cmd /
WORKDIR /src
COPY . .
CMD ["/cmd"]
