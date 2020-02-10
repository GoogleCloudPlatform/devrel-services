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

WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services
COPY . .

# BEGIN Jank to get an upstream patch into the build
WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services
RUN git clone https://go.googlesource.com/build build
WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services/build
RUN git fetch https://go.googlesource.com/build refs/changes/38/176638/5 && git checkout FETCH_HEAD

RUN echo "replace golang.org/x/build => ../build " >> ../drghs-worker/go.mod

WORKDIR /src/github.com/GoogleCloudPlatform/devrel-services/drghs-worker

# END Jank

RUN go install github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd

FROM alpine as health

RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

FROM golang:1.11

COPY --from=build /go/bin/maintnerd /
COPY --from=health /bin/grpc_health_probe /bin/grpc_health_probe

WORKDIR /src
COPY . .
CMD ["/maintnerd"]
