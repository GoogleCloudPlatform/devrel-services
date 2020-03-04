package internalapi

import (
	"context"
	"fmt"
	"strings"

	maintner_internal "github.com/GoogleCloudPlatform/devrel-services/drghs-worker/internal"
	"golang.org/x/build/maintner"
)

// TransferProxyServer implements maintner_internal.InternalIssueServiceServer
type TransferProxyServer struct {
	c *maintner.Corpus
}

var _ maintner_internal.InternalIssueServiceServer = &TransferProxyServer{}

// NewTransferProxyServer builds and returns a TransferProxyServer
func NewTransferProxyServer(c *maintner.Corpus) *TransferProxyServer {
	return &TransferProxyServer{
		c: c,
	}
}

// TombstoneIssues tombstones the requested issues that are in the corpus
func (s *TransferProxyServer) TombstoneIssues(ctx context.Context, r *maintner_internal.TombstoneIssuesRequest) (*maintner_internal.TombstoneIssuesResponse, error) {
	var ntombstoned int32

	err := s.c.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := getRepoPath(repo)
		if !strings.HasPrefix(r.Parent, repoID) {
			// Not our repository... ignore
			fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Parent)
			return nil
		}

		for _, iss := range r.IssueNumbers {
			fmt.Printf("for repository %v requested to tombstone issue: %v", r.Parent, iss)
			// go through the corpus, find this repo, find the issue, tombstone it

			issue := repo.Issue(iss)
			if issue == nil {
				return fmt.Errorf("Issue: %v not found", iss)
			}
			err := repo.MarkTombstoned(issue)
			if err != nil {
				return err
			}
			ntombstoned++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &maintner_internal.TombstoneIssuesResponse{
		TombstonedCount: ntombstoned,
	}, nil
}

func getRepoPath(ta *maintner.GitHubRepo) string {
	return fmt.Sprintf("%v/%v", ta.ID().Owner, ta.ID().Repo)
}
