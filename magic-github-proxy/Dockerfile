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

FROM python:3.7.1-slim-stretch as base

RUN apt update &&  \
    apt install -y \
    wget \
    tar \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists*

WORKDIR /work/theacodes/magic-github-proxy
RUN wget -O - https://github.com/theacodes/magic-github-proxy/tarball/master | tar xzC . --strip-components 1
WORKDIR /work/theacodes/magic-github-proxy
RUN pip install -r requirements.txt
RUN python3 setup.py install

WORKDIR /app/magic-github-proxy

COPY . .

RUN pip install --upgrade google-cloud-storage
RUN pip install --upgrade google-cloud-kms
RUN pip install --upgrade flask
RUN pip install --upgrade requests


CMD ["python3", "main.py"]
