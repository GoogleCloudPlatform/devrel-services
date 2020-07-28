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

package filter

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
)

const defaultFilter = "true"

// Owner checks if the Owner passes the given CEL expression.
func Owner(o *drghs_v1.Owner, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}
	if o == nil {
		return false, nil
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Owner{}),
		cel.Declarations(
			decls.NewIdent("owner", decls.NewObjectType("drghs.v1.Owner"), nil)))

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := prg.Eval(map[string]interface{}{
		"owner": o,
	})

	return out == types.True, err
}

// Repository checks if the Repository passes the given CEL expression.
func Repository(r *drghs_v1.Repository, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}
	if r == nil {
		return false, nil
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Repository{}),
		cel.Declarations(
			decls.NewIdent("repository", decls.NewObjectType("drghs.v1.Repository"), nil)))

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := prg.Eval(map[string]interface{}{
		"repository": r,
	})

	return out == types.True, err
}

// Slo checks if the Slo passes the given CEL expression.
func Slo(s *drghs_v1.SLO, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}
	if s == nil {
		return false, nil
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.SLO{}),
		cel.Declarations(
			decls.NewIdent("slo", decls.NewObjectType("drghs.v1.SLO"), nil)))

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := prg.Eval(map[string]interface{}{
		"slo": s,
	})

	return out == types.True, err
}
