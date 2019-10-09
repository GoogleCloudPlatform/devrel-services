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

package apiroutes

import (
	"devrel/cloud/devrel-github-service/drghs-worker/pkg/googlers"
	"devrel/cloud/devrel-github-service/drghs-worker/pkg/status"
	"devrel/cloud/devrel-github-service/drghs-worker/pkg/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/build/maintner"
)

type vOneAPI struct {
	corpus          *maintner.Corpus
	googlerResolver googlers.GooglersResolver
	route           *mux.Router
}

func (api vOneAPI) Routes() {
	api.route.HandleFunc("/{owner}/{repository}/issues", api.handleListIssues()).Methods("GET")
	api.route.HandleFunc("/{owner}/{repository}/issues/{issue:[0-9]+}", api.handleGetIssue()).Methods("GET")
	api.route.HandleFunc("/approvedPRs", api.handleGetApprovedPRs()).Methods("GET")
	api.route.HandleFunc("/sloViolations", api.handleGetSloViolations()).Methods("POST")
}

// NewV1Api registers api v1 routes
func NewV1Api(cor *maintner.Corpus, resolver googlers.GooglersResolver, route *mux.Router) (ApiRoute, error) {
	if cor == nil {
		return nil, fmt.Errorf("corpus must not be nil")
	}
	if route == nil {
		return nil, fmt.Errorf("route must not be nil")
	}

	api := &vOneAPI{
		corpus:          cor,
		googlerResolver: resolver,
		route:           route,
	}

	return api, nil
}

func (api vOneAPI) handleListIssues() http.HandlerFunc {
	corpus := api.corpus

	return func(w http.ResponseWriter, r *http.Request) {
		// Get the repository this is for
		vars := mux.Vars(r)
		rOwner := vars["owner"]
		rRepo := vars["repository"]

		var rPullRequest *bool
		var rClosed *bool
		rIncludeComments := false
		rIncludeReviews := false

		// Get our querystring parameters
		if param := r.URL.Query().Get("pull_request"); param != "" {
			shouldIncludePR, err := strconv.ParseBool(param)
			if err == nil {
				rPullRequest = &shouldIncludePR
			}
		}

		if param := r.URL.Query().Get("reviews"); param != "" {
			shouldIncludeReviews, err := strconv.ParseBool(param)
			if err == nil {
				rIncludeReviews = shouldIncludeReviews
			}
		}

		if param := r.URL.Query().Get("comments"); param != "" {
			shouldIncludeComments, err := strconv.ParseBool(param)
			if err == nil {
				rIncludeComments = shouldIncludeComments
			}
		}

		if param := r.URL.Query().Get("closed"); param != "" {
			shouldIncludeClosed, err := strconv.ParseBool(param)
			if err == nil {
				rClosed = &shouldIncludeClosed
			}
		}

		resp, err := makeIssuesResponse(corpus, rOwner+"/"+rRepo, rPullRequest, rClosed, rIncludeComments, rIncludeReviews)

		if err != nil {
			return
		}

		resp.WriteTo(w)
	}
}

func (api vOneAPI) handleGetIssue() http.HandlerFunc {

	corpus := api.corpus

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		owner := vars["owner"]
		repository := vars["repository"]
		rIssueID, err := strconv.Atoi(vars["issue"])

		if err != nil {
			return
		}

		rIncludeComments := true
		rIncludeReviews := true

		if param := r.URL.Query().Get("comments"); param != "" {
			shouldIncludeComments, err := strconv.ParseBool(param)
			if err == nil {
				rIncludeComments = shouldIncludeComments
			}
		}

		if param := r.URL.Query().Get("reviews"); param != "" {
			shouldIncludeReviews, err := strconv.ParseBool(param)
			if err == nil {
				rIncludeReviews = shouldIncludeReviews
			}
		}

		var resp status.Response

		corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := repo.ID().String()
			if (owner + "/" + repository) != repoID {
				// Ignore this repo.
				return nil
			}

			return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
				if issue.NotExist || resp.Issue != nil {
					return nil
				}

				issueID := int(issue.Number)
				if rIssueID != issueID {
					// Ignore this issue.
					return nil
				}

				s := utils.TranslateIssueToStatus(issue, repoID, rIncludeComments, rIncludeReviews)

				resp.Issue = s
				return nil
			})
		})

		resp.WriteTo(w)
	}
}

func (api vOneAPI) handleGetApprovedPRs() http.HandlerFunc {
	return commonGetApprovedPrs(api.corpus)
}

func (api vOneAPI) handleGetSloViolations() http.HandlerFunc {
	return commonGetSloViolations(api.corpus, api.googlerResolver)
}
