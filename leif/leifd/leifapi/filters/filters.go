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

// BuildOwnerFilter builds a CEL program to filter an Owner with the given CEL expression
func BuildOwnerFilter(filter string) (cel.Program, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Owner{}),
		cel.Declarations(
			decls.NewIdent("owner", decls.NewObjectType("drghs.v1.Owner"), nil)))

	if err != nil {
		return nil, err
	}

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	return env.Program(checked)
}

// Owner checks if the given Owner passes the given CEL program
func Owner(o *drghs_v1.Owner, p cel.Program) (bool, error) {
	if o == nil || p == nil {
		return false, nil
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := p.Eval(map[string]interface{}{
		"owner": o,
	})

	return out == types.True, err
}

// BuildRepositoryFilter builds a CEL program to filter a Repo with the given CEL expression
func BuildRepositoryFilter(filter string) (cel.Program, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Repository{}),
		cel.Declarations(
			decls.NewIdent("repository", decls.NewObjectType("drghs.v1.Repository"), nil)))

	if err != nil {
		return nil, err
	}

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	return env.Program(checked)
}

// Repository checks if the Repository passes the given CEL expression
func Repository(r *drghs_v1.Repository, p cel.Program) (bool, error) {
	if r == nil || p == nil {
		return false, nil
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := p.Eval(map[string]interface{}{
		"repository": r,
	})

	return out == types.True, err
}

// BuildSloFilter builds a CEL program to filter an Slo with the given CEL expression
func BuildSloFilter(filter string) (cel.Program, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.SLO{}),
		cel.Declarations(
			decls.NewIdent("slo", decls.NewObjectType("drghs.v1.SLO"), nil)))

	if err != nil {
		return nil, err
	}

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	return env.Program(checked)
}

// Slo checks if the Slo passes the given CEL expression.
func Slo(s *drghs_v1.SLO, p cel.Program) (bool, error) {
	if s == nil || p == nil {
		return false, nil
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := p.Eval(map[string]interface{}{
		"slo": s,
	})

	return out == types.True, err
}
