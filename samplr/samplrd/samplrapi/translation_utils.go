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
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/samplr"

	"github.com/golang/protobuf/ptypes"
)

func makeRepositoryPB(rname string) (*drghs_v1.Repository, error) {
	return &drghs_v1.Repository{
		Name: rname,
	}, nil
}

func makeSnippetPB(snippet *samplr.Snippet) (*drghs_v1.Snippet, error) {
	versionPB, err := makeSnippetVersionPB(snippet.Primary)
	if err != nil {
		return nil, err
	}
	return &drghs_v1.Snippet{
		Name:     snippet.Name,
		Language: snippet.Language,
		Primary:  versionPB,
	}, nil
}

func makeSnippetVersionPB(version samplr.SnippetVersion) (*drghs_v1.SnippetVersion, error) {
	filePb, err := makeFilePB(version.File)
	if err != nil {
		return nil, err
	}
	metaPb, err := makeSnippetVersionMetaPB(version.Meta)
	if err != nil {
		return nil, err
	}

	return &drghs_v1.SnippetVersion{
		Name:    version.Name,
		File:    filePb,
		Lines:   version.Lines,
		Content: version.Content,
		Meta:    metaPb,
	}, nil
}

func makeSnippetVersionMetaPB(meta samplr.SnippetVersionMeta) (*drghs_v1.SnippetVersionMeta, error) {
	return &drghs_v1.SnippetVersionMeta{
		Title:       meta.Title,
		Description: meta.Description,
		Usage:       meta.Usage,
		ApiVersion:  meta.ApiVersion,
	}, nil
}

func makeFilePB(file *samplr.File) (*drghs_v1.File, error) {
	commitPb, err := makeGitCommitPB(file.GitCommit)
	if err != nil {
		return nil, err
	}
	return &drghs_v1.File{
		Filepath:  file.FilePath,
		GitCommit: commitPb,
		Size:      int32(file.Size),
	}, nil
}

func makeGitCommitPB(commit *samplr.GitCommit) (*drghs_v1.GitCommit, error) {
	authoredTime, err := ptypes.TimestampProto(commit.AuthoredTime)
	if err != nil {
		return nil, err
	}
	committedTime, err := ptypes.TimestampProto(commit.CommittedTime)
	if err != nil {
		return nil, err
	}
	return &drghs_v1.GitCommit{
		Name:           commit.Name,
		Subject:        commit.Subject,
		AuthorEmail:    commit.AuthorEmail,
		AuthoredTime:   authoredTime,
		CommitterEmail: commit.CommitterEmail,
		CommittedTime:  committedTime,
		Sha:            commit.Hash,
	}, nil
}

func makeSnippetProtoversion(snippets []*samplr.Snippet) []*drghs_v1.Snippet {
	protoversions := make([]*drghs_v1.Snippet, len(snippets))
	for idx, s := range snippets {
		protoversions[idx], _ = makeSnippetPB(s)
	}
	return protoversions
}

func makeSnippetVersionProtoversion(versions []samplr.SnippetVersion) []*drghs_v1.SnippetVersion {
	protoversions := make([]*drghs_v1.SnippetVersion, len(versions))
	for idx, v := range versions {
		protoversions[idx], _ = makeSnippetVersionPB(v)
	}
	return protoversions
}
