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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/rtr"
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"cloud.google.com/go/errorreporting"
	"github.com/sirupsen/logrus"
)

var (
	listen      = flag.String("listen", ":6343", "listen address")
	verbose     = flag.Bool("verbose", false, "enable verbose debug output")
	rmQuery     = flag.String("rmquery", "", "query to remove from the URL before proxying. e.g. 'key'")
	supervisor  = flag.String("sprvsr", "", "the name of the service that is hosting the supervisor")
	errorClient *errorreporting.Client
	pathRegex   = regexp.MustCompile(`^\/api\/v1\/([\w-]+)\/([\w-]+)\/issues[\w\/-]*$`)
)

const (
	// Using a reserved TLD https://tools.ietf.org/html/rfc2606
	devnull = "devnull.invalid"
)

// Log
var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}

	log.Out = os.Stdout
}

func main() {
	flag.Parse()

	if *verbose {
		log.Level = logrus.TraceLevel
	}

	if *supervisor == "" {
		log.Fatal("error: must specify --supervisor")
	}

	if *listen == "" {
		log.Fatal("error: must specify --listen")
	}

	rtr.ListenAndServe(*listen, *supervisor, []string{}, []string{*rmQuery}, calculateHost, log)
}

func calculateHost(req *http.Request) (string, error) {
	path := req.URL.Path
	// We might need to put some more real "smarts" to this logic
	// in the event we need to handle the /v1/owners/*/repositories
	// call, which asks for a list of all repositories in a given org.
	// We might need to call out to a different API, but for now we can
	// forward to "null"?

	if path == "/update" {
		return devnull, nil
	}

	// As of right now, this function assumes all calls into the
	// proxy are of form /v1/owners/OWNERNAME/repositories/REPOSITORYNAME/issues/*
	log.Tracef("Matching path againtst regex: %v", path)
	mtches := pathRegex.FindAllStringSubmatch(path, -1)
	if mtches != nil {
		log.Tracef("have a v1 path: %v", path)
		// This match will be of form:
		// [["/v1/owners/foo/repositories/bar1/issues" "foo" "bar1"]]
		// Therefore slice the array

		ta := repos.TrackedRepository{
			Owner: mtches[0][1],
			Name: mtches[0][2],
		}

		sn, err := serviceName(ta)
		if err != nil {
			return "", err
		}
		return sn, nil
	} else if strings.HasPrefix(path, "/api/v0") {
		log.Tracef("have a v0 path: %v", path)
		// Need to check the body (json) which will contain the owner
		// and repository information
		if req.Method != http.MethodPost {
			return "", fmt.Errorf("api/v0 must be POST http methods. Got: %v", req.Method)
		}
		// Parse body
		var dat map[string]interface{}

		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		// This is because ioutil.ReadAll closes the body
		// which will break the reverse-proxy
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		if err := json.Unmarshal(bodyBytes, &dat); err != nil {
			return "", err
		}

		repo := ""
		// Grab the repository
		if d, ok := dat["repo"]; ok {
			repo = d.(string)
		}
		if d, ok := dat["Repo"]; ok && repo == "" {
			repo = d.(string)
		}

		if repo == "" {
			return "", fmt.Errorf("did not specify repository in body")
		}

		parts := strings.Split(repo,"/")
		if len(parts) != 2 {
			return "", fmt.Errorf("bad format for repo")
		}

		ta := repos.TrackedRepository{
			Owner: parts[0],
			Name: parts[1],
		}

		sn, err := serviceName(ta)
		if err != nil {
			return "", err
		}
		return sn, nil
	}
	return devnull, nil
}

func serviceName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-s-%s", t.RepoSha())), nil
}