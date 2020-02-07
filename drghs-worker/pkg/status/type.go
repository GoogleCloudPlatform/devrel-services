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

package status

// Type indicates the Type of Issue or PR it is
type Type string

const (
	// TypeBug indicates a Bugs
	TypeBug = Type("Bug")
	// TypeFeature indicates a feature
	TypeFeature = Type("FR")
	// TypeCleanup indicates cleanup process.
	TypeCleanup = Type("Cleanup")
	// TypeCustomer indicates a customer request
	TypeCustomer = Type("Customer")
	// TypeProcess indicates a proceess issue.
	TypeProcess = Type("Process")
	// TypePR indicates a pull request
	TypePR = Type("PR")
	// TypeUnknown indicates no type could be determined
	TypeUnknown = Type("")
)
