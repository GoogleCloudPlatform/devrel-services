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

// Priority is used to describe an SLO's criticality.
// An SLO is defined in US business days (i.e. weekends, US Holidays excluded) for any SLO that is 7 days or less. For any SLO that is 8+ days, it is defined in calendar days
// P0 15-30 minute response, ASAP resolution
// P1 1 day response, 1 week resolution
// P2 5 day response, best effort resolution
// P3 6 month response, best effort resolution
// P4 12 month response, best effort resolution
type Priority uint

const (
	// P0 is the highest priority
	P0 = Priority(iota)
	// P1 is the second higest priority
	P1
	// P2 is the third highest priority
	P2
	// P3 is the fourth highest priority
	P3
	// P4 is the fifth highest priority
	P4
)
