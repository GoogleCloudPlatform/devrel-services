package main

import (
	"testing"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/google/go-cmp/cmp"
)

func TestRepoToTrackedRepo(t *testing.T) {
	cs := []struct {
		Repo drghs_v1.Repository
		Want *repos.TrackedRepository
		Name string
	}{
		{
			Name: "Expected",
			Repo: drghs_v1.Repository{
				Name: "GoogleCloudPlatform/devrel-services",
			},
			Want: &repos.TrackedRepository{
				Owner: "GoogleCloudPlatform",
				Name: "devrel-services",
			},
		},
		{
			Name: "Failure",
			Repo: drghs_v1.Repository{
				Name: "owners/GoogleCloudPlatform/repos/devrel-services",
			},
			Want: nil,
		},
	}
	for _, c := range cs {
		got := repoToTrackedRepo(&c.Repo)
		if diff := cmp.Diff(c.Want, got); diff != "" {
			t.Errorf("Test: %v Repositories differ (-want +got)\n%s", c.Name, diff)
		}
	}
}
