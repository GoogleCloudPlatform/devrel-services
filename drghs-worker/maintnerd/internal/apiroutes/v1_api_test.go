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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"golang.org/x/build/maintner"
)

type googlerVoid struct {
}

func (gv googlerVoid) IsGoogler(user string) bool { return true }
func (gv googlerVoid) Update()                    {}

func TestV1HandlesApprovedPRs(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("GET", "/approvedPRs", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestV1FailsSloViolationsBody(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/sloViolations", strings.NewReader("RANDOM"))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestV1SloViolationsNoConfigs(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/sloViolations", strings.NewReader("{\"Configs\":[]}"))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestV1HandlesGetIssues(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("GET", "/GoogleCloudPlatform/google-cloud-node/issues", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestV1HandlesGetIssue(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("GET", "/GoogleCloudPlatform/google-cloud-node/issues/1234", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestV1GetIssueMustBeDigits(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV1Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("GET", "/GoogleCloudPlatform/google-cloud-node/issues/1234asdf", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusNotFound)
}
