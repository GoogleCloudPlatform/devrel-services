package main

import (
	"net/http"
	"testing"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/time/rate"
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
				Name:  "devrel-services",
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

type mockRoundTripper struct {
	Response *http.Response
	Err      error
}

func (t mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.Response, t.Err
}

func TestLimitTransportLimits(t *testing.T) {
	mqps := 1
	limit := rate.Every(time.Second / time.Duration(mqps))
	limitr := rate.NewLimiter(limit, mqps)
	mrtr := mockRoundTripper{
		Response: nil,
		Err:      nil,
	}

	lt := limitTransport{
		limiter: limitr,
		base:    mrtr,
	}

	tst := time.Now()
	max := 4
	for i := 0; i < max; i++ {
		r, err := lt.RoundTrip(&http.Request{})
		if r != mrtr.Response {
			t.Errorf("limitTransport did not return expected response")
		}
		if err != mrtr.Err {
			t.Errorf("limitTransport did not return expected error")
		}
	}

	got := time.Since(tst).Seconds()
	want := float64((max - 1) * mqps)
	if got < want {
		t.Errorf("limit transport did not limit our qps appropriately\nWant: %v Got: %v", want, got)
	}
}

func TestLimitTransportHandlesNil(t *testing.T) {
	mrtr := mockRoundTripper{
		Response: nil,
		Err:      nil,
	}

	lt := limitTransport{
		limiter: nil,
		base:    mrtr,
	}

	max := 4
	for i := 0; i < max; i++ {
		r, err := lt.RoundTrip(&http.Request{})
		if r != mrtr.Response {
			t.Errorf("limitTransport did not return expected response")
		}
		if err != mrtr.Err {
			t.Errorf("limitTransport did not return expected error")
		}
	}
}
