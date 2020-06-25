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

package leif

import (
	"context"
	"errors"
	"fmt"
	"time"

	goGH "github.com/google/go-github/v32/github"
)

const SLOConfigPath = "issue_slo_rules.json"

// SLORule represents a service level objective (SLO) rule
type SLORule struct {
	AppliesTo          AppliesTo          `json:"appliesTo"`
	ComplianceSettings ComplianceSettings `json:"complianceSettings"`
}

// AppliesTo stores structured data on which issues and/or pull requests a SLO applies to
type AppliesTo struct {
	GitHubLabels         []string `json:"gitHubLabels"`
	ExcludedGitHubLabels []string `json:"excludedGitHubLabels"`
	Issues               bool     `json:"issues"`
	PRs                  bool     `json:"prs"`
}

// ComplianceSetting stores data on the requirements for an issue or pull request to be considered compliant with the SLO
type ComplianceSettings struct {
	ResponseTime     time.Duration `json:"responseTime"`
	ResolutionTime   time.Duration `json:"resolutionTime"`
	RequiresAssignee bool          `json:"requiresAssignee"`
	Responders       Responders    `json:"responders"`
}

// Responders stores structured data on the responders to the issue or pull request the SLO applies to
type Responders struct {
	Owners       []string `json:"owners"`
	Contributors string   `json:"contributors"`
	Users        []string `json:"users"`
}

type Repository struct {
	SLOfile  *goGH.RepositoryContent
	SLORules []*SLORule
}

func (repo *Repository) FindRepository(ctx context.Context, reponame string) error {

	// ts := oauth2.StaticTokenSource(
	// 	&oauth2.Token{AccessToken: "acctok"},
	// )
	// tc := oauth2.NewClient(ctx, ts)
	client := goGH.NewClient(nil)

	// org, _, _ := client.Organizations.Get(ctx, "google")

	// url := org.GetReposURL()
	// client.Repositories.Get

	// meh, r, err := client.Repositories.ListByOrg(ctx, "google", nil)

	// fmt.Println(meh)
	// fmt.Println(r)
	// fmt.Println(err)

	// var e *goGH.RateLimitError
	// file, _, resp, err := client.Repositories.GetContents(ctx, "BrennaEpp", "quasar", ".github/CDE_OF_CONDUCT.md", nil)

	// if errors.As(err, &e) {
	// 	fmt.Println("error")
	// }
	// if file == nil {
	// 	//look at org
	// 	//issue_slo_rules.json
	// 	file, dirCont, resp, err = client.Repositories.GetContents(ctx, "Google", ".github", "CONTRIBUTING.md", nil)
	// 	if file == nil {
	// 		fmt.Println(err)
	// 		return nil
	// 	}

	// }
	// fmt.Println(repCont)
	// org, _, err := client.Organizations.Get(ctx, "Google")

	// repos, _, error := client.Repositories.ListByOrg(ctx, "GoogleCloudPlatform", nil)
	// fmt.Println(error)
	// for i, repo := range repos {
	// 	fmt.Print(i)
	// 	fmt.Println(repo.GetName())
	// }

	fmt.Println("meow")

	// fmt.Println(resp)
	// fmt.Println(err)
	// fmt.Println(reflect.TypeOf(err))
	// fmt.Println(file)
	// fmt.Println(repCont.Encoding)
	// fmt.Println(file.GetContent())

	// a, err := file.GetContent()
	// fmt.Println(reflect.TypeOf(a))
	// fmt.Println(err)

	rep := Repository{}
	err := rep.findSLODoc(ctx, "google", "blockly", client)
	fmt.Println(err)
	err = rep.parseSLOs()
	fmt.Println(err)

	return nil
}

func (repo *Repository) parseSLOs() error {
	if repo.SLOfile == nil {
		return errors.New("Repository has no SLO rules config")
	}

	file, err := repo.SLOfile.GetContent()
	if err != nil {
		return err
	}

	slos, err := unmarshalSLOs([]byte(file))
	if err != nil {
		return err
	}

	repo.SLORules = slos

	return nil
}

func (repo *Repository) findSLODoc(ctx context.Context, orgName string, repoName string, ghClient *goGH.Client) error {
	var ghErrorResponse *goGH.ErrorResponse

	content, _, _, err := ghClient.Repositories.GetContents(ctx, orgName, repoName, ".github/"+SLOConfigPath, nil)

	if errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404 {
		// SLO config not found, look for file in org:
		content, _, _, err = ghClient.Repositories.GetContents(ctx, orgName, ".github", SLOConfigPath, nil)
	}
	if err != nil {
		return err
	}

	repo.SLOfile = content

	return nil
}
