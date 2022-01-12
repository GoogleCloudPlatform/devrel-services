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

FROM golang:1.17 as build

ENV GO111MODULE=on

WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services
COPY . .

WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services/drghs-worker/issue-tombstoner

RUN go install github.com/GoogleCloudPlatform/devrel-services/drghs-worker/issue-tombstoner

FROM debian:latest
COPY --from=build /go/bin/issue-tombstoner /bin
WORKDIR /src
COPY . .
ENTRYPOINT ["/bin/bash"]
