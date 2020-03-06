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

	cases := []struct {
		QPS        float64
		Iterations int
	}{
		{
			QPS:        .5,
			Iterations: 4,
		},
		{
			QPS:        1,
			Iterations: 3,
		},
		{
			QPS:        2,
			Iterations: 5,
		},
		{
			QPS:        1,
			Iterations: 1,
		},
		{
			QPS:        2,
			Iterations: 1,
		},
	}

	for _, c := range cases {
		dur := time.Duration(float64(time.Second) * (1.0 / c.QPS))
		limit := rate.Every(dur)
		limiter := rate.NewLimiter(limit, 1)
		mrtr := mockRoundTripper{
			Response: nil,
			Err:      nil,
		}

		lt := limitTransport{
			limiter: limiter,
			base:    mrtr,
		}

		tst := time.Now()

		for i := 0; i < c.Iterations; i++ {
			r, err := lt.RoundTrip(&http.Request{})
			if r != mrtr.Response {
				t.Errorf("limitTransport did not return expected response")
			}
			if err != mrtr.Err {
				t.Errorf("limitTransport did not return expected error. got: %v", err)
			}
		}

		got := time.Since(tst)
		want := time.Duration(float64(time.Second) * ((1.0 / c.QPS) * float64(c.Iterations-1)))
		if got < want {
			t.Errorf("limit transport did not limit our qps appropriately\nWant: %v Got: %v", want, got)
		}
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

func TestBuildLimiter(t *testing.T) {
	cases := []struct {
		Name      string
		NIssues   int32
		WantLimit rate.Limit
	}{
		{
			Name:      "LowFrequency",
			NIssues:   864000,
			WantLimit: 0.1,
		},
		{
			Name:      "HighFrequency",
			NIssues:   86400000,
			WantLimit: 10,
		},
		{
			Name:      "EvenFrequency",
			NIssues:   8640000,
			WantLimit: 1,
		},
	}
	for _, c := range cases {

		limiter := buildLimiter(c.NIssues)
		gotLimit := limiter.Limit()
		if gotLimit != c.WantLimit {
			t.Errorf("test: %v failed. limiter improperly set. got: %v want: %v", c.Name, gotLimit, c.WantLimit)
		}
	}
}
