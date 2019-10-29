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

package samplrapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrd/filter"

	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
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

type SampleServiceServer struct {
	c   *samplr.Corpus
	gp  *gitCommitPaginator
	sp  *snippetPaginator
	svp *snippetVersionPaginator
	tr  *trackedRepositoryPaginator
}

func NewSampleServiceServer(c *samplr.Corpus) *SampleServiceServer {
	return &SampleServiceServer{
		c: c,
		sp: &snippetPaginator{
			set: make(map[time.Time]snippetPage),
		},
		svp: &snippetVersionPaginator{
			set: make(map[time.Time]snippetVersionPage),
		},
		gp: &gitCommitPaginator{
			set: make(map[time.Time]gitCommitPage),
		},
		tr: &trackedRepositoryPaginator{
			set: make(map[time.Time]trackedRepositoryPage),
		},
	}
}

func (s *SampleServiceServer) ListRepositories(ctx context.Context, req *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {
	var pg []string
	var idx int
	var err error
	nextToken := ""

	re := regexp.MustCompile(`owners/([\w-_]+|\*)`)
	if !re.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	if req.PageToken != "" {
		//Handle pagination
		pageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		pg, idx, err = s.tr.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)
		if err != nil {
			return nil, err
		}
	} else {
		repos := make([]string, 0)

		parts := re.FindStringSubmatch(req.Parent)
		owner := parts[1]

		filter := func(w samplr.WatchedRepository) bool {
			if w.Owner() != owner {
				return false
			}
			return true
		}
		if owner == "*" {
			filter = func(w samplr.WatchedRepository) bool {
				return true
			}
		}

		err := s.c.ForEachRepoF(func(w samplr.WatchedRepository) error {
			repos = append(repos, fmt.Sprintf("owners/%v/repositories/%v", w.Owner(), w.RepositoryName()))
			return nil
		}, filter)

		// Create Page
		t, err := s.tr.CreatePage(repos)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		//Get page
		pg, idx, err = s.tr.GetPage(t, pagesize)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if idx > 0 {
			nextToken, err = makeFirstPageToken(t, idx)
			if err != nil {
				return nil, err
			}
		}
	}

	protorepositories := make([]*drghs_v1.Repository, 0)
	for _, v := range pg {
		pr, err := makeRepositoryPB(v)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		should, err := filter.FilterRepository(pr, req.Filter)
		if err != nil {
			log.Errorf("Issue filtering repository: %v", err)
			return nil, err
		}

		if should {
			protorepositories = append(protorepositories, pr)
		}
	}

	return &drghs_v1.ListRepositoriesResponse{
		Repositories:  protorepositories,
		NextPageToken: nextToken,
		Total:         int32(len(pg)),
	}, err
}

func (s *SampleServiceServer) ListGitCommits(ctx context.Context, req *drghs_v1.ListGitCommitsRequest) (*drghs_v1.ListGitCommitsResponse, error) {
	var pg []*samplr.GitCommit
	var idx int
	var err error
	nextToken := ""

	re := regexp.MustCompile(`owners/([\w-_]+)/repositories/([\w-_]+)`)

	if !re.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	if req.PageToken != "" {
		//Handle pagination
		pageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		pg, idx, err = s.gp.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)
		if err != nil {
			return nil, err
		}
	} else {
		parts := re.FindStringSubmatch(req.Parent)
		owner := parts[1]
		repo := parts[2]

		commits := make([]*samplr.GitCommit, 0)

		filter := func(w samplr.WatchedRepository) bool {
			if w.Owner() != owner || w.RepositoryName() != repo {
				return false
			}
			return true
		}

		s.c.ForEachRepoF(func(watchedRepo samplr.WatchedRepository) error {
			err := watchedRepo.ForEachGitCommit(func(commit *samplr.GitCommit) error {
				commits = append(commits, commit)
				return nil
			})
			return err
		}, filter)

		// Create Page
		t, err := s.gp.CreatePage(commits)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		//Get page
		pg, idx, err = s.gp.GetPage(t, pagesize)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if idx > 0 {
			nextToken, err = makeFirstPageToken(t, idx)
			if err != nil {
				return nil, err
			}
		}
	}

	protocommits := make([]*drghs_v1.GitCommit, 0)
	for _, v := range pg {
		pc, err := makeGitCommitPB(v)
		if err != nil {
			log.Errorf("Could not get commit pb %v", err)
			return nil, err
		}

		should, err := filter.FilterGitCommit(pc, req.Filter)
		if err != nil {
			log.Errorf("Issue filtering commit: %v", err)
			return nil, err
		}

		if should {
			protocommits = append(protocommits, pc)
		}
	}

	log.Infof("len of protocommits: %v", len(protocommits))

	log.Warnf("Err: %v", err)
	return &drghs_v1.ListGitCommitsResponse{
		GitCommits:    protocommits,
		NextPageToken: nextToken,
		Total:         int32(len(protocommits)),
	}, err
}

func (s *SampleServiceServer) GetGitCommit(ctx context.Context, req *drghs_v1.GetGitCommitRequest) (*drghs_v1.GitCommit, error) {
	return &drghs_v1.GitCommit{}, nil
}

func (s *SampleServiceServer) ListFiles(ctx context.Context, req *drghs_v1.ListFilesRequest) (*drghs_v1.ListFilesResponse, error) {
	return &drghs_v1.ListFilesResponse{}, nil
}

func (s *SampleServiceServer) ListSnippets(ctx context.Context, req *drghs_v1.ListSnippetsRequest) (*drghs_v1.ListSnippetsResponse, error) {
	var pg []*samplr.Snippet
	var idx int
	var err error
	nextToken := ""

	re := regexp.MustCompile(`owners/([\w-_]+)/repositories/([\w-_]+)`)

	if !re.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	if req.PageToken != "" {
		//Handle pagination
		pageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		pg, idx, err = s.sp.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)
		if err != nil {
			return nil, err
		}
	} else {
		parts := re.FindStringSubmatch(req.Parent)
		owner := parts[1]
		repo := parts[2]

		snippets := make([]*samplr.Snippet, 0)

		repositoryFilter := func(w samplr.WatchedRepository) bool {
			if w.Owner() != owner || w.RepositoryName() != repo {
				return false
			}
			return true
		}

		s.c.ForEachRepoF(func(watchedRepo samplr.WatchedRepository) error {
			err := watchedRepo.ForEachSnippet(func(snippet *samplr.Snippet) error {
				snippets = append(snippets, snippet)
				return nil
			})
			return err
		}, repositoryFilter)

		// Filter Snippets
		filteredSnippets := make([]*samplr.Snippet, 0)
		for _, v := range snippets {
			pv, err := makeSnippetPB(v)
			if err != nil {
				log.Errorf("Could not get version pb %v", err)
				return nil, err
			}

			should, err := filter.FilterSnippet(pv, req.Filter)
			if err != nil {
				log.Errorf("Issue filtering repository: %v", err)
				return nil, err
			}

			if should {
				filteredSnippets = append(filteredSnippets, v)
			}
		}

		// Create Page
		t, err := s.sp.CreatePage(filteredSnippets)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		//Get page
		pg, idx, err = s.sp.GetPage(t, pagesize)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if idx > 0 {
			nextToken, err = makeFirstPageToken(t, idx)
			if err != nil {
				return nil, err
			}
		}
	}

	protoversions, err := makeSnippetProtoversion(pg)
	if err != nil {
		log.Errorf("Could not create the snippet protoversion: %v", err)
		return nil, err
	}

	return &drghs_v1.ListSnippetsResponse{
		Snippets:      protoversions,
		NextPageToken: nextToken,
		Total:         int32(len(protoversions)),
	}, err
}

func (s *SampleServiceServer) ListSnippetVersions(ctx context.Context, req *drghs_v1.ListSnippetVersionsRequest) (*drghs_v1.ListSnippetVersionsResponse, error) {
	var pg []samplr.SnippetVersion
	var idx int
	var err error
	nextToken := ""

	re := regexp.MustCompile(`owners/([\w-_]+)/repositories/([\w-_]+)/snippets/([\w-_]+)/languages/([\w-_]+)`)

	if !re.MatchString(req.Parent) {
		log.Errorf("Regex failed to match: %v", req.Parent)
		return nil, errors.New("Could not gather route data")
	}

	if req.PageToken != "" {
		//Handle pagination
		pageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		pg, idx, err = s.svp.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)
		if err != nil {
			return nil, err
		}

	} else {

		parts := re.FindStringSubmatch(req.Parent)
		owner := parts[1]
		repo := parts[2]

		versions := make([]samplr.SnippetVersion, 0)

		parentFilter := func(w samplr.WatchedRepository) bool {
			if w.Owner() != owner || w.RepositoryName() != repo {
				return false
			}
			return true
		}

		snippetFilter := func(s *samplr.Snippet) bool {
			if s.Name != req.Parent {
				return false
			}
			return true
		}

		err := s.c.ForEachRepoF(func(watchedRepo samplr.WatchedRepository) error {
			err := watchedRepo.ForEachSnippetF(func(snippet *samplr.Snippet) error {
				for _, version := range snippet.Versions {
					versions = append(versions, version)
				}
				return nil
			}, snippetFilter)

			return err
		}, parentFilter)

		// Filter Snippet version
		filteredVersions := make([]samplr.SnippetVersion, 0)
		for _, v := range versions {
			pv, err := makeSnippetVersionPB(v)
			if err != nil {
				log.Errorf("Could not get version pb %v", err)
				return nil, err
			}

			should, err := filter.FilterSnippetVersion(pv, req.Filter)
			if err != nil {
				log.Errorf("Issue filtering snippet version: %v", err)
				return nil, err
			}

			if should {
				filteredVersions = append(filteredVersions, v)
			}
		}

		// Create Page
		t, err := s.svp.CreatePage(filteredVersions)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		pagesize := getPageSize(int(req.PageSize))

		//Get page
		pg, idx, err = s.svp.GetPage(t, pagesize)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if idx > 0 {
			tsp, err := ptypes.TimestampProto(t)
			if err != nil {
				log.Errorf("Could not make timestamp %v", err)
				return nil, err
			}
			nextToken, err = makeNextPageToken(&drghs_v1.PageToken{
				FirstRequestTimeUsec: tsp,
				Offset:               int32(idx),
			}, idx)
			if err != nil {
				return nil, err
			}
		}
	}

	protoversions, err := makeSnippetVersionProtoversion(pg)
	if err != nil {
		log.Errorf("Could not create the snippet version protoversion: %v", err)
		return nil, err
	}
	return &drghs_v1.ListSnippetVersionsResponse{
		SnippetVersions: protoversions,
		NextPageToken:   nextToken,
		Total:           int32(len(protoversions)),
	}, err
}
