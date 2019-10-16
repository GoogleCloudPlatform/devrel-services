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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"golang.org/x/build/maintner"
)

type vZeroAPI struct {
	corpus          *maintner.Corpus
	googlerResolver googlers.GooglersResolver
	route           *mux.Router
}

func (api vZeroAPI) Routes() {
	api.route.HandleFunc("/issue", api.handleGetIssue())
	api.route.HandleFunc("/issues", api.handleListIssues())
	api.route.HandleFunc("/sloViolations", api.handleSloViolantions())
	api.route.HandleFunc("/approvedPRs", api.handleApprovedPrs())
}

// NewV0Api registers api v0 routes
func NewV0Api(cor *maintner.Corpus, resolver googlers.GooglersResolver, route *mux.Router) (ApiRoute, error) {
	if cor == nil {
		return nil, fmt.Errorf("corpus must not be nil")
	}
	if route == nil {
		return nil, fmt.Errorf("route must not be nil")
	}

	apiHandler := vZeroAPI{
		corpus:          cor,
		googlerResolver: resolver,
		route:           route,
	}

	return apiHandler, nil
}

func (api vZeroAPI) handleGetIssue() http.HandlerFunc {
	corpus := api.corpus

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			status.Response{Error: "must use POST"}.WriteTo(w)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			status.Response{Error: err.Error()}.WriteTo(w)
			return
		}

		var req status.GetIssueRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			status.Response{Error: err.Error()}.WriteTo(w)
			return
		}

		var resp status.Response

		corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := repo.ID().String()
			if req.Repo != repoID {
				// Ignore this repo.
				return nil
			}

			return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
				if issue.NotExist || resp.Issue != nil {
					return nil
				}

				issueID := issue.Number
				if req.Issue != int(issueID) {
					// Ignore this issue.
					return nil
				}

				s := utils.TranslateIssueToStatus(issue, repoID, req.IncludeComments, req.IncludeReviews)

				resp.Issue = s
				return nil
			})
		})

		resp.WriteTo(w)
	}
}

func (api vZeroAPI) handleListIssues() http.HandlerFunc {
	corpus := api.corpus

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			status.Response{Error: "must use POST"}.WriteTo(w)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			status.Response{Error: err.Error()}.WriteTo(w)
			return
		}

		var req status.ListIssuesRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			status.Response{Error: err.Error()}.WriteTo(w)
			return
		}

		PullRequest := utils.UnmarshalBool(req.PullRequest)
		Closed := utils.UnmarshalBool(req.Closed)

		resp, err := makeIssuesResponse(corpus, req.Repo, PullRequest, Closed, req.IncludeComments, req.IncludeReviews)

		if err != nil {
			return
		}

		resp.WriteTo(w)
	}
}

func (api vZeroAPI) handleSloViolantions() http.HandlerFunc {
	return commonGetSloViolations(api.corpus, api.googlerResolver)
}

func (api vZeroAPI) handleApprovedPrs() http.HandlerFunc {
	return commonGetApprovedPrs(api.corpus)
}
