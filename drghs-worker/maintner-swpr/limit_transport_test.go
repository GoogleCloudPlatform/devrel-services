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

package main

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

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
