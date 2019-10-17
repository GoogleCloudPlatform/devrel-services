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
	"io/ioutil"
	"net/http"

	"golang.org/x/build/maintner"
)

func makeIssuesResponse(corpus *maintner.Corpus, rRepoID string, incPr *bool, incClosed *bool, incComments bool, incReviews bool) (status.Response, error) {
	var resp status.Response

	corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := repo.ID().String()
		if rRepoID != repoID {
			// Ignore this repo.
			return nil
		}

		return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
			if issue.NotExist {
				return nil
			}

			if incPr != nil && issue.PullRequest != *incPr {
				return nil
			}
			if incClosed != nil && issue.Closed != *incClosed {
				return nil
			}

			s := utils.TranslateIssueToStatus(issue, rRepoID, incComments, incReviews)

			resp.Issues = append(resp.Issues, s)
			return nil
		})
	})

	return resp, nil
}

func commonGetApprovedPrs(cor *maintner.Corpus) http.HandlerFunc {
	corpus := cor

	return func(w http.ResponseWriter, r *http.Request) {
		var resp status.Response

		corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := repo.ID().String()
			return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
				if issue.NotExist || !issue.PullRequest || !utils.IsApproved(issue) {
					return nil
				}

				s := utils.TranslateIssueToStatus(issue, repoID, true, true)

				resp.Issues = append(resp.Issues, s)
				return nil
			})
		})

		resp.WriteTo(w)
	}
}

func commonGetSloViolations(corp *maintner.Corpus, resolver googlers.GooglersResolver) http.HandlerFunc {
	corpus := corp
	googlerResolver := resolver

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

		var req status.Request
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			status.Response{Error: err.Error()}.WriteTo(w)
			return
		}

		// repo ID -> configs
		configs := make(map[string]*status.RequestConfig)
		for _, conf := range req.Configs {
			for _, repo := range conf.Repos {
				if configs[repo] != nil {
					w.WriteHeader(http.StatusBadRequest)
					status.Response{Error: "repo had more than one SLO config applied"}.WriteTo(w)
					return
				}
				configs[repo] = &conf
			}
		}

		var resp status.Response

		corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := repo.ID().String()
			slo := configs[repoID]
			if slo == nil {
				// Ignore this repo.
				return nil
			}

			return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
				if issue.Closed || issue.NotExist || issue.PullRequest {
					return nil
				}

				userLast := issue.Created
				googlerLast := issue.Created
				issue.ForeachComment(func(comment *maintner.GitHubComment) error {
					googler := googlerResolver.IsGoogler(comment.User.Login)
					if googler {
						googlerLast = comment.Created
					} else {
						userLast = comment.Created
					}
					return nil
				})

				s := &status.Status{
					Issue:             issue,
					Repo:              repoID,
					Priority:          status.P2,
					PriorityUnknown:   true,
					Created:           issue.Created,
					LastGooglerUpdate: googlerLast,
					LastUserUpdate:    userLast,
				}

				s.Closed = false
				s.FillWithSLO(slo)

				resp.Issues = append(resp.Issues, s)
				return nil
			})
		})

		resp.WriteTo(w)
	}
}
