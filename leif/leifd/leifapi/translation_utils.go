// Copyright 2020 Google LLC
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

package leifapi

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
)

func makeRepositoryPB(rname string) (*drghs_v1.Repository, error) {
	return &drghs_v1.Repository{
		Name: rname,
	}, nil
}

func makeOwnerPB(name string) (*drghs_v1.Owner, error) {
	return &drghs_v1.Owner{
		Name: name,
	}, nil
}
