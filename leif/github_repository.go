package leif

import (
	"context"
	"errors"
	"fmt"

	goGH "github.com/google/go-github/v32/github"
)

const SLOConfigPath = "issue_slo_rules.json"

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
