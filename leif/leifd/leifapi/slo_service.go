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
	paginator "github.com/GoogleCloudPlatform/devrel-services/leif/leifd/leifapi/pagination"

	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var reposParent = regexp.MustCompile(`owners/([\w-_]+|\*)`)
var slosParent = regexp.MustCompile(`owners/([\w-_]+|\*)/repositories/([\w-_]+|\*)`)

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
	ownerPaginator *paginator.Strings
	repoPaginator  *paginator.Strings
	sloPaginator   *paginator.Slo
}

// NewSLOServiceServer builds and returns a new SLOServiceServer
func NewSLOServiceServer(c *leif.Corpus) *SLOServiceServer {
	s := &SLOServiceServer{
		c:              c,
		ownerPaginator: &paginator.Strings{Log: log},
		repoPaginator:  &paginator.Strings{Log: log},
		sloPaginator:   &paginator.Slo{Log: log},
	}

	s.ownerPaginator.Init()
	s.repoPaginator.Init()
	s.sloPaginator.Init()

	return s
}

// ListOwners returns the list of Owners tracked by the Corpus
func (s *SLOServiceServer) ListOwners(ctx context.Context, req *drghs_v1.ListOwnersRequest) (*drghs_v1.ListOwnersResponse, error) {
	owners, nextToken, err := s.handleOwnerPagination(req.PageToken, req.PageSize, req.OrderBy)
	if err != nil {
		return nil, err
	}

	filterP, err := filter.BuildOwnerFilter(req.Filter)
	if err != nil {
		return nil, err
	}

	protoOwners := make([]*drghs_v1.Owner, 0)
	for _, o := range owners {
		protoOwner, err := makeOwnerPB(o)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Owner(protoOwner, filterP)
		if err != nil {
			log.Errorf("Issue filtering owner: %v", err)
			return nil, err
		}

		if include {
			protoOwners = append(protoOwners, protoOwner)
		}
	}

	return &drghs_v1.ListOwnersResponse{
		Owners:        protoOwners,
		NextPageToken: nextToken,
		Total:         int32(len(owners)),
	}, err
}

func (s *SLOServiceServer) handleOwnerPagination(pToken string, pSize int32, orderBy string) ([]string, string, error) {
	var pg []string
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := paginator.DecodePageToken(pToken)
		if err != nil {
			return nil, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		pg, index, err = s.ownerPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, "", err
		}
		nextToken, err = paginator.MakeNextPageToken(pageToken, index)
	} else {
		owners := make([]string, 0)

		if orderBy == "" {
			err = s.c.ForEachOwner(func(o leif.Owner) error {
				owners = append(owners, fmt.Sprintf("owners/%v", o.Name()))
				return nil
			})
		} else {
			var sortFn func(o []*leif.Owner) func(i, j int) bool

			switch orderBy {
			case "name":
				sortFn = func(o []*leif.Owner) func(i, j int) bool {
					return func(i, j int) bool { return o[i].Name() < o[j].Name() }
				}
			case "-name":
				sortFn = func(o []*leif.Owner) func(i, j int) bool {
					return func(i, j int) bool { return o[i].Name() > o[j].Name() }
				}
			default:
				return nil, "", fmt.Errorf("Cannot order repositories by %s", orderBy)
			}

			err = s.c.ForEachOwnerFSort(
				func(o leif.Owner) error {
					owners = append(owners, fmt.Sprintf("owners/%v", o.Name()))
					return nil
				},
				func(o leif.Owner) bool { return true },
				sortFn,
			)
		}
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		// Create Page
		t, err := s.ownerPaginator.CreatePage(owners)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		//Get page
		pg, index, err = s.ownerPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		if index > 0 {
			nextToken, err = paginator.MakeFirstPageToken(t, index)
		}
	}
	return pg, nextToken, err
}

// ListRepositories returns the list of Repositories tracked by the Corpus
func (s *SLOServiceServer) ListRepositories(ctx context.Context, req *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {

	if !reposParent.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	repos, nextToken, err := s.handleRepoPagination(req.PageToken, req.PageSize, req.OrderBy, req.Parent)
	if err != nil {
		return nil, err
	}

	filterP, err := filter.BuildRepositoryFilter(req.Filter)
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

		include, err := filter.Repository(protoRepo, filterP)
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

func (s *SLOServiceServer) handleRepoPagination(pToken string, pSize int32, orderBy string, parent string) ([]string, string, error) {
	var pg []string
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := paginator.DecodePageToken(pToken)
		if err != nil {
			return nil, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		pg, index, err = s.repoPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, "", err
		}
		nextToken, err = paginator.MakeNextPageToken(pageToken, index)
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

		if orderBy == "" {
			err = s.c.ForEachRepoF(func(repo leif.Repository) error {
				repos = append(repos, fmt.Sprintf("owners/%v/repositories/%v", repo.OwnerName(), repo.RepoName()))
				return nil
			}, filter)
		} else {
			var sortFn func(repos []*leif.Repository) func(i, j int) bool

			switch orderBy {
			case "name":
				sortFn = func(repos []*leif.Repository) func(i, j int) bool {
					return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
				}
			case "-name":
				sortFn = func(repos []*leif.Repository) func(i, j int) bool {
					return func(i, j int) bool { return repos[i].RepoName() > repos[j].RepoName() }
				}
			default:
				return nil, "", fmt.Errorf("Cannot order repositories by %s", orderBy)
			}

			err = s.c.ForEachRepoFSort(
				func(repo leif.Repository) error {
					repos = append(repos, fmt.Sprintf("owners/%v/repositories/%v", repo.OwnerName(), repo.RepoName()))
					return nil
				},
				filter,
				sortFn,
			)
		}
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		// Create Page
		t, err := s.repoPaginator.CreatePage(repos)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		//Get page
		pg, index, err = s.repoPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		if index > 0 {
			nextToken, err = paginator.MakeFirstPageToken(t, index)
		}
	}
	return pg, nextToken, err
}

// ListOwnerSLOs returns the list of slos for an owner tracked by the Corpus
func (s *SLOServiceServer) ListOwnerSLOs(ctx context.Context, req *drghs_v1.ListSLOsRequest) (*drghs_v1.ListSLOsResponse, error) {

	if !reposParent.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	slos, nextToken, err := s.handleOwnerSloPagination(req.PageToken, req.PageSize, req.Parent)
	if err != nil {
		return nil, err
	}

	filterP, err := filter.BuildSloFilter(req.Filter)
	if err != nil {
		return nil, err
	}

	protoSlos := make([]*drghs_v1.SLO, 0)
	for _, slo := range slos {
		protoSlo, err := makeSloPB(slo)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Slo(protoSlo, filterP)
		if err != nil {
			log.Errorf("Issue filtering repository: %v", err)
			return nil, err
		}

		if include {
			protoSlos = append(protoSlos, protoSlo)
		}
	}

	return &drghs_v1.ListSLOsResponse{
		Slos:          protoSlos,
		NextPageToken: nextToken,
		Total:         int32(len(slos)),
	}, err
}

// ListSLOs returns the list of slos for a repository tracked by the Corpus
func (s *SLOServiceServer) ListSLOs(ctx context.Context, req *drghs_v1.ListSLOsRequest) (*drghs_v1.ListSLOsResponse, error) {

	if !slosParent.MatchString(req.Parent) {
		return nil, fmt.Errorf("Invalid parent: %v", req.Parent)
	}

	slos, nextToken, err := s.handleRepoSloPagination(req.PageToken, req.PageSize, req.Parent)
	if err != nil {
		return nil, err
	}

	filterP, err := filter.BuildSloFilter(req.Filter)
	if err != nil {
		return nil, err
	}

	protoSlos := make([]*drghs_v1.SLO, 0)
	for _, slo := range slos {
		protoSlo, err := makeSloPB(slo)
		if err != nil {
			log.Errorf("Could not create repository pb %v", err)
			return nil, err
		}

		include, err := filter.Slo(protoSlo, filterP)
		if err != nil {
			log.Errorf("Issue filtering repository: %v", err)
			return nil, err
		}

		if include {
			protoSlos = append(protoSlos, protoSlo)
		}
	}

	return &drghs_v1.ListSLOsResponse{
		Slos:          protoSlos,
		NextPageToken: nextToken,
		Total:         int32(len(slos)),
	}, err
}

func (s *SLOServiceServer) handleOwnerSloPagination(pToken string, pSize int32, parent string) ([]*leif.SLORule, string, error) {
	var pg []*leif.SLORule
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := paginator.DecodePageToken(pToken)
		if err != nil {
			return nil, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		pg, index, err = s.sloPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, "", err
		}
		nextToken, err = paginator.MakeNextPageToken(pageToken, index)
	} else {
		slos := make([]*leif.SLORule, 0)

		parts := reposParent.FindStringSubmatch(parent)

		owner := parts[1]

		parentFilter := func(o leif.Owner) bool {
			return o.Name() == owner
		}

		if owner == "*" {
			parentFilter = func(o leif.Owner) bool {
				return true
			}
		}

		err = s.c.ForEachOwnerF(func(o leif.Owner) error {
			slos = append(slos, o.SLORules...)
			return nil
		}, parentFilter)

		// Create Page
		t, err := s.sloPaginator.CreatePage(slos)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		//Get page
		pg, index, err = s.sloPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		if index > 0 {
			nextToken, err = paginator.MakeFirstPageToken(t, index)
		}
	}
	return pg, nextToken, err
}

func (s *SLOServiceServer) handleRepoSloPagination(pToken string, pSize int32, parent string) ([]*leif.SLORule, string, error) {
	var pg []*leif.SLORule
	var index int
	var err error
	nextToken := ""

	if pToken != "" {
		pageToken, err := paginator.DecodePageToken(pToken)
		if err != nil {
			return nil, "", err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			log.Errorf("Error deserializing time: %v", err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		pg, index, err = s.sloPaginator.GetPage(ftime, pageSize)
		if err != nil {
			return nil, "", err
		}
		nextToken, err = paginator.MakeNextPageToken(pageToken, index)
	} else {
		slos := make([]*leif.SLORule, 0)

		parts := slosParent.FindStringSubmatch(parent)

		repo := parts[2]
		owner := parts[1]

		parentFilter := func(r leif.Repository) bool {
			return r.OwnerName() == owner && r.RepoName() == repo
		}

		if repo == "*" {
			parentFilter = func(r leif.Repository) bool {
				return r.OwnerName() == owner
			}
		}
		if owner == "*" {
			parentFilter = func(r leif.Repository) bool {
				return r.RepoName() == repo
			}
		}
		if owner == "*" && repo == "*" {
			parentFilter = func(r leif.Repository) bool {
				return true
			}
		}

		err = s.c.ForEachRepoF(func(repo leif.Repository) error {
			slos = append(slos, repo.SLORules...)
			return nil
		}, parentFilter)

		// Create Page
		t, err := s.sloPaginator.CreatePage(slos)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		pageSize := paginator.GetPageSize(int(pSize))

		//Get page
		pg, index, err = s.sloPaginator.GetPage(t, pageSize)
		if err != nil {
			log.Error(err)
			return nil, "", err
		}

		if index > 0 {
			nextToken, err = paginator.MakeFirstPageToken(t, index)
		}
	}
	return pg, nextToken, err
}
