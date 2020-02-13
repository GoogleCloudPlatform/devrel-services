package internalapi

import (
	"fmt"
	"io"
	"strings"

	maintner_internal "github.com/GoogleCloudPlatform/devrel-services/drghs-worker/internal"
	"golang.org/x/build/maintner"
)

// TransferProxyServer implements maintner_internal.InternalIssueServiceServer
type TransferProxyServer struct {
	c *maintner.Corpus
}

var errDidTransfer error = fmt.Errorf("Transfered issue")

var _ maintner_internal.InternalIssueServiceServer = &TransferProxyServer{}

// NewTransferProxyServer builds and returns a TransferProxyServer
func NewTransferProxyServer(c *maintner.Corpus) *TransferProxyServer {
	return &TransferProxyServer{
		c: c,
	}
}

// TombstoneIssues tombstones the requested issues that are in the corpus
func (s *TransferProxyServer) TombstoneIssues(stream maintner_internal.InternalIssueService_TombstoneIssuesServer) error {
	var ntombstoned int32

	for {
		iss, err := stream.Recv()
		if err == io.EOF {
			// Done. Close and Return
			return stream.SendAndClose(&maintner_internal.TombstoneIssueResponse{
				NumberTombstoned: ntombstoned,
			})
		}
		if err != nil {
			return err
		}

		fmt.Printf("for repository %v requested to tombstone issue: %v", iss.Parent, iss.IssueNumber)
		// go through the corpus, find this repo, find the issue, tombstone it

		err = s.c.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := getRepoPath(repo)
			if !strings.HasPrefix(iss.Parent, repoID) {
				// Not our repository... ignore
				fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, iss.Parent)
				return nil
			}

			issue := repo.Issue(iss.IssueNumber)
			if issue == nil {
				return fmt.Errorf("Issue: %v not found", iss.IssueNumber)
			}
			err := repo.MarkTombstoned(issue)
			if err == nil {
				err = errDidTransfer
			}
			return err
		})
		if err != nil && err == errDidTransfer {
			ntombstoned++
			err = nil
		}
		return err
	}
}

func getRepoPath(ta *maintner.GitHubRepo) string {
	return fmt.Sprintf("%v/%v", ta.ID().Owner, ta.ID().Repo)
}
