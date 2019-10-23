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

package rtr

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

// HostCalculator takes a given request and calculates
// the host to resolve to.
type HostCalculator func(*http.Request) (string, error)

// DEVNULL is a DNS name that is guarenteed to fail a DNS lookup
// due to using a reserved TLD https://tools.ietf.org/html/rfc2606
const DEVNULL = "devnull.invalid"

// ListenAndServe listens on the provided address and Reverse-Proxies incoming requests to the endpoints as
// calculated by hostCalculator, while removing the headers and queries specified.
func ListenAndServe(addr string, supervisorHost string, headersToRemove, queriesToRemove []string, hostCalculator HostCalculator, log *logrus.Logger) error {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Got an error calling hostname: %v", err)
		return err
	}

	sprvsrurl, err := url.Parse(supervisorHost)
	if err != nil {
		log.Fatalf("error parsing supervisor url: %v", err)
		return err
	}

	// Need to reverse-proxy all our requests to fan-out
	reverseProxy := &httputil.ReverseProxy{}
	reverseProxy.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", hostname)
		// For now hardcoding the scheme
		req.URL.Scheme = "http"

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		query := req.URL.Query()
		for _, q := range queriesToRemove {
			if _, ok := query[q]; ok {
				query.Del(q)
			}
		}
		req.URL.RawQuery = query.Encode()

		headers := req.Header
		for _, h := range headersToRemove {
			if _, ok := headers[h]; ok {
				headers.Del(h)
			}
		}
		req.Header = headers

		// Rewrite the Host based on the path
		host, err := hostCalculator(req)
		if err != nil {
			log.Errorf("Encountered error calculating host. Rewriting to %v. Error: %v", DEVNULL, err)
		}
		req.URL.Host = host
		log.Tracef("Rewrote host to: %v", req.URL.Host)
	}
	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		log.Debugf("Got response status: %v", resp.StatusCode)
		log.Debugf("Got body len: %v", resp.ContentLength)
		return nil
	}

	rpsprvsr := httputil.NewSingleHostReverseProxy(sprvsrurl)
	rpsprvsr.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", hostname)
		// For now hardcoding the scheme
		req.URL.Scheme = "http"

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		req.URL.Host = supervisorHost
		// This application is behind Cloud Endpoints which uses
		// a Query parameter named "key" to authenticate. We need to strip
		// that custom key from our URL before proxying
		query := req.URL.Query()
		for _, q := range queriesToRemove {
			if _, ok := query[q]; ok {
				query.Del(q)
			}
		}
		req.URL.RawQuery = query.Encode()

		headers := req.Header
		for _, h := range headersToRemove {
			if _, ok := headers[h]; ok {
				headers.Del(h)
			}
		}
		req.Header = headers
	}

	// Send everything through Mux
	r := mux.NewRouter()
	r.Handle("/update", rpsprvsr)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	r.PathPrefix("/").Handler(reverseProxy)

	// Add middleware support
	n := negroni.New()
	l := negroni.NewLogger()
	// Custom logger for Negroni as otherwise it will log log the API key
	l.SetFormat("{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}}")
	n.Use(l)
	n.Use(negroni.NewRecovery())
	n.UseHandler(r)

	return http.ListenAndServe(addr, n)
}
