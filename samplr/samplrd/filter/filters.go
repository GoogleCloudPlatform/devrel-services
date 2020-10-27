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

package filter

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
)

const defaultFilter = "true"

// BuildSnippetFilter creates a cel Program based off of the given filter string
func BuildSnippetFilter(filter string) (cel.Program, error) {

	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Snippet{}),
		cel.Declarations(
			decls.NewIdent("snippet", decls.NewObjectType("drghs.v1.Snippet"), nil),
		))

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

// Snippet checks if the Snippet passes the given CEL program.
func Snippet(s *drghs_v1.Snippet, p cel.Program) (bool, error) {
	if s == nil || p == nil {
		return false, nil
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := p.Eval(map[string]interface{}{
		"snippet": s,
	})

	return out == types.True, err
}

// SnippetVersion checks if the SnippetVersion passes the given CEL expression.
func SnippetVersion(s *drghs_v1.SnippetVersion, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.SnippetVersion{}),
		cel.Declarations(
			decls.NewIdent("version", decls.NewObjectType("drghs.v1.SnippetVersion"), nil)))

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
		"version": s,
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

// GitCommit checks if the GitCommit passes the given CEL expression.
func GitCommit(g *drghs_v1.GitCommit, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.GitCommit{}),
		cel.Declarations(
			decls.NewIdent("commit", decls.NewObjectType("drghs.v1.GitCommit"), nil)))

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
		"commit": g,
	})

	return out == types.True, err
}
