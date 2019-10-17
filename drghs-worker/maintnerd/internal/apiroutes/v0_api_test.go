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

	"golang.org/x/build/maintner"

	"github.com/matryer/is"
)

func TestHandlesApprovedPRs(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV0Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("GET", "/approvedPRs", nil)
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestFailsSloViolationsBody(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV0Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/sloViolations", strings.NewReader("RANDOM"))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestSloViolationsNoConfigs(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV0Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/sloViolations", strings.NewReader("{\"Configs\":[]}"))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestHandlesGetIssues(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV0Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/issues", strings.NewReader(`{"Repo": "gcp/gcp"}`))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}

func TestHandlesGetIssue(t *testing.T) {
	is := is.New(t)

	cor := &maintner.Corpus{}
	resolver := googlerVoid{}
	router := mux.NewRouter()

	srv, err := NewV0Api(cor, resolver, router)

	is.NoErr(err)

	// Hook up the routes
	srv.Routes()

	req, err := http.NewRequest("POST", "/issue", strings.NewReader(`{"Repo": "gcp/gcp","Issue": 1234}`))
	is.NoErr(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
}
