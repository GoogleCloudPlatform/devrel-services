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

package filters

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
)

const defaultFilter = "true"

// FilterIssue determines if an Issue satisfies the constraints in the
// ListIssuesRequest
func FilterIssue(issue *drghs_v1.Issue, r *drghs_v1.ListIssuesRequest) (bool, error) {
	switch x := r.PullRequestNullable.(type) {
	case *drghs_v1.ListIssuesRequest_PullRequest:
		if issue.IsPr != x.PullRequest {
			return false, nil
		}
	case nil:
		// Do nothing
	default:
		// Do nothing
	}

	switch x := r.ClosedNullable.(type) {
	case *drghs_v1.ListIssuesRequest_Closed:
		if issue.Closed != x.Closed {
			return false, nil
		}
	case nil:
		// Do nothing
	default:
		// Do nothing
	}

	return true, nil
}

// FilterRepository determines if a Repository matches the CEL spec
// for the given filter
func FilterRepository(r *drghs_v1.Repository, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Repository{}),
		cel.Declarations(decls.NewIdent("repository", decls.NewObjectType("drghs.v1.Repository"), nil)))

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

// FilterComment determines if a GitHubComment matches the CEL spec
// for the given filter
func FilterComment(c *drghs_v1.GitHubComment, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Types(&drghs_v1.Issue{}),
		cel.Declarations(decls.NewIdent("comment", decls.NewObjectType("drghs.v1.GitHubComment"), nil)))

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
		"comment": c,
	})

	return out == types.True, err
}
