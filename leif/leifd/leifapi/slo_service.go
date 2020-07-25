// Copyright 2020 Google LLC
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

package leifapi

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/leif"
	filter "github.com/GoogleCloudPlatform/devrel-services/leif/leifd/leifapi/filters"

	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var reposParent = regexp.MustCompile(`owners/([\w-_]+|\*)`)
var slosParent = regexp.MustCompile(`owners/([\w-_]+)/repositories/([\w-_]+)`)

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

var _ drghs_v1.SLOServiceServer = &SLOServiceServer{}

// SLOServiceServer is an implementation of drghs_v1.SLOServiceServer
type SLOServiceServer struct {
	c              *leif.Corpus
	ownerPaginator *stringsPaginator
	repoPaginator  *stringsPaginator
	sloPaginator   *paginator
}

// NewSLOServiceServer builds and returns a new SLOServiceServer
func NewSLOServiceServer(c *leif.Corpus) *SLOServiceServer {
	return &SLOServiceServer{
		c: c,
		ownerPaginator: &stringsPaginator{
			set: make(map[time.Time]stringsPage),
		},
		repoPaginator: &stringsPaginator{
			set: make(map[time.Time]stringsPage),
		},
		sloPaginator: &paginator{
			set: make(map[time.Time]page),
		},
	}
}

// ListOwners returns the list of Owners tracked by the Corpus
func (s *SLOServiceServer) ListOwners(ctx context.Context, req *drghs_v1.ListOwnersRequest) (*drghs_v1.ListOwnersResponse, error) {

	owners, index, nextToken, err := s.handleOwnerPagination(req.PageToken, req.PageSize)
	if err != nil {
		return nil, err
	}

	protoOwners := make([]*drghs_v1.Owner, 0)
	for _, o := range owners {
		protoO, err := makeOwnerPB(o)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Owner(protoO, req.Filter)
		if err != nil {
			log.Errorf("Issue filtering owner: %v", err)
			return nil, err
		}

		if include {
			protoOwners = append(protoOwners, protoO)
		}
	}

	return &drghs_v1.ListOwnersResponse{
		Owners:        protoOwners,
		NextPageToken: nextToken,
		Total:         int32(len(repos)),
	}, err
}

func (s *SLOServiceServer) handleOwnerPagination(pToken string, pSize int32) ([]string, int, string, error) {
	var pg []string
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := decodePageToken(pToken)
		if err != nil {
			return nil, -1, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		pg, index, err = s.repoPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, -1, "", err
		}
		nextToken, err := makeNextPageToken(pageToken, index)
	} else {
		owners := make([]string, 0)

		err := s.c.ForEachOwnerF(func(o leif.Owner) error {
			owners = append(owners, fmt.Sprintf("owners %v/", o.Name()))
			return nil
		}, filter)

		// Create Page
		t, err := s.ownerPaginator.CreatePage(owners)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		//Get page
		pg, index, err := s.ownerPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		if index > 0 {
			nextToken, err := makeFirstPageToken(t, index)
		}
	}
	return pg, index, nextToken, err
}

// ListRepositories returns the list of Repositories tracked by the Corpus
func (s *SLOServiceServer) ListRepositories(ctx context.Context, req *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {

	if !reposParent.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	repos, index, nextToken, err := s.handleRepoPagination(req.PageToken, req.PageSize, req.Parent)
	if err != nil {
		return nil, err
	}

	protoRepositories := make([]*drghs_v1.Repository, 0)
	for _, repo := range repos {
		protoRepo, err := makeRepositoryPB(repo)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Repository(protoRepo, req.Filter)
		if err != nil {
			log.Errorf("Issue filtering repository: %v", err)
			return nil, err
		}

		if include {
			protoRepositories = append(protoRepositories, protoRepo)
		}
	}

	return &drghs_v1.ListRepositoriesResponse{
		Repositories:  protoRepositories,
		NextPageToken: nextToken,
		Total:         int32(len(repos)),
	}, err
}

func (s *SLOServiceServer) handleRepoPagination(pToken string, pSize int32, parent string) ([]string, int, string, error) {
	var pg []string
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := decodePageToken(pToken)
		if err != nil {
			return nil, -1, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		pg, index, err = s.repoPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, -1, "", err
		}
		nextToken, err := makeNextPageToken(pageToken, index)
	} else {
		repos := make([]string, 0)

		parts := reposParent.FindStringSubmatch(parent)
		owner := parts[1]

		filter := func(repo leif.Repository) bool {
			return repo.OwnerName() == owner
		}
		if owner == "*" {
			filter = func(repo leif.Repository) bool {
				return true
			}
		}

		err := s.c.ForEachRepoF(func(repo leif.Repository) error {
			repos = append(repos, fmt.Sprintf("owners/%v/repositories/%v", repo.OwnerName(), repo.RepoName()))
			return nil
		}, filter)

		// Create Page
		t, err := s.repoPaginator.CreatePage(repos)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		//Get page
		pg, index, err := s.repoPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		if index > 0 {
			nextToken, err := makeFirstPageToken(t, index)
		}
	}
	return pg, index, nextToken, err
}

// ListSLOs returns the list of slos for a repository tracked by the Corpus
func (s *SLOServiceServer) ListSLOs(ctx context.Context, req *drghs_v1.ListSLOsRequest) (*drghs_v1.ListSLOsResponse, error) {

	if !slosParent.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	slos, index, nextToken, err := s.handleSloPagination(req.PageToken, req.PageSize, req.Parent)
	if err != nil {
		return nil, err
	}

	protoSlos := make([]*drghs_v1.SLO, 0)
	for _, slo := range repos {
		protoRepo, err := makeRepositoryPB(repo)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Repository(protoRepo, req.Filter)
		if err != nil {
			log.Errorf("Issue filtering repository: %v", err)
			return nil, err
		}

		if include {
			protoRepositories = append(protoRepositories, protoRepo)
		}
	}

	return &drghs_v1.ListRepositoriesResponse{
		Repositories:  protoRepositories,
		NextPageToken: nextToken,
		Total:         int32(len(repos)),
	}, err
}

func (s *SLOServiceServer) handleSloPagination(pToken string, pSize int32, parent string) ([]*leif.SLORule, int, string, error) {
	var pg []*leif.SLORule
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := decodePageToken(pToken)
		if err != nil {
			return nil, -1, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		pg, index, err = s.repoPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, -1, "", err
		}
		nextToken, err := makeNextPageToken(pageToken, index)
	} else {
		repos := make([]string, 0)

		parts := reposParent.FindStringSubmatch(parent)
		owner := parts[1]

		filter := func(repo leif.Repository) bool {
			return repo.OwnerName() == owner
		}
		if owner == "*" {
			filter = func(repo leif.Repository) bool {
				return true
			}
		}

		err := s.c.ForEachRepoF(func(repo leif.Repository) error {
			repos = append(repos, fmt.Sprintf("owners/%v/repositories/%v", repo.OwnerName(), repo.RepoName()))
			return nil
		}, filter)

		// Create Page
		t, err := s.repoPaginator.CreatePage(repos)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		pageSize := getPageSize(int(pSize))

		//Get page
		pg, index, err := s.repoPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, -1, "", err
		}

		if index > 0 {
			nextToken, err := makeFirstPageToken(t, index)
		}
	}
	return pg, index, nextToken, err
}
