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

package git

import (
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	// log.Level = logrus.DebugLevel
	log.Out = os.Stdout
}

func VerboseLog() {
	log.Level = logrus.DebugLevel
}

func FormatLog(f logrus.Formatter) {
	log.Formatter = f
}

const (
	// GitDirName is the name of the directory storing Git data
	GitDirName = ".git"
)

var (
	// NoErrAlreadyUpToDate an error stating the repository is already up to date
	NoErrAlreadyUpToDate = errors.New("already up to date")
	// ErrBranchExists an error stating the specified branch already exists
	ErrBranchExists = errors.New("branch already exists")
	// ErrBranchNotFound an error stating the specified branch does not exist
	ErrBranchNotFound = errors.New("branch not found")
	// ErrTagExists an error stating the specified tag already exists
	ErrTagExists = errors.New("tag already exists")
	// ErrTagNotFound an error stating the specified tag does not exist
	ErrTagNotFound = errors.New("tag not found")
	// ErrFetching is returned when the packfile could not be downloaded
	ErrFetching = errors.New("unable to fetch packfile")

	ErrInvalidReference          = errors.New("invalid reference, should be a tag or a branch")
	ErrRepositoryNotExists       = errors.New("repository does not exist")
	ErrRepositoryAlreadyExists   = errors.New("repository already exists")
	ErrRemoteNotFound            = errors.New("remote not found")
	ErrRemoteExists              = errors.New("remote already exists")
	ErrWorktreeNotProvided       = errors.New("worktree should be provided")
	ErrIsBareRepository          = errors.New("worktree not available in a bare repository")
	ErrUnableToResolveCommit     = errors.New("unable to resolve commit")
	ErrPackedObjectsNotSupported = errors.New("Packed objects not supported")

	//ErrStop is used to stop a ForEach function in an Iter
	ErrStop = errors.New("stop iter")
)

// PlainOpen returns a Repository from the provided path
func PlainOpen(directory string) (repository *Repository, err error) {
	_, err = os.Stat(directory)
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(path.Join(directory, GitDirName)); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrRepositoryNotExists
		}
		return nil, err
	}
	return &Repository{dir: directory}, nil
}

// PlainClone clones the repository specified into the given directory
func PlainClone(directory string, plain bool, options *CloneOptions) (repository *Repository, err error) {
	if options == nil {
		return nil, errors.New("Must provide some options")
	}

	url := options.URL

	cloneCmd := exec.Command("git", "clone", "--single-branch", url, directory)
	_, err = cloneCmd.Output()
	if err != nil {
		return nil, err
	}

	return &Repository{dir: directory, url: url}, nil
}

// CloneOptions stores options for cloning a Repository
type CloneOptions struct {
	URL string
}
