#!/bin/bash

# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eo pipefail
set -x

export GO111MODULE=on # Always use modules.
TIMEOUT=45m

go version
date

# Re-organize files
export GOPATH=$PWD/gopath
target=$GOPATH/src/github.com/GoogleCloudPlatform
mkdir -p $target
mv github/devrel-services $target
cd $target/devrel-services

testdirs=( repos rtr samplr sprvsr drghs-worker )

OUTFILE=$(pwd)/gotest.out
for t in "${testdirs[@]}"; do
    # Subshell to avoid having to `cd ..`
    (
    cd "$t" || exit 1
    echo "$t"
    go test -timeout $TIMEOUT -v ./... 2>&1 | tee -a $OUTFILE

	if [ $GOLANG_SAMPLES_GO_VET ]; then
        diff -u <(echo -n) <(gofmt -d -s .)
		# We are cd'd in the directory. Simply go vet ./...
        go vet ./...
    fi
    )
done

# Do the easy stuff before running tests. Fail fast!

date

cat $OUTFILE > sponge_log.xml
