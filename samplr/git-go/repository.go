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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	commitFormat = "--pretty=commit: %H%nAuthor: %an%nAuthor Email: %ae%nAuthor Date: %at%nCommitter: %cn%nCommitter Email: %ce%nCommitter Date: %ct%nSubject: %s%n"
)

var (
	commitItemRegExp = regexp.MustCompile(`(?m)Commit: (\w{40})\nAuthor: (.*)\nAuthor Email: (.*)\nAuthor Date: (.*)\nCommitter: (.*)\nCommitter Email: (.*)\nCommitter Date: (.*)\nSubject: (.*)\nBody:(.*)\nFiles:\n((^[ADRMC](\d{3})?\s+[\w-\.\/]+(\s+[\w-\.\/]+)?\n)+)*`)
	fileNameRegex    = regexp.MustCompile(`^([\w\/\.-]+)$`)
	refsRegex        = regexp.MustCompile(`(\w{40}) ([\w\/\-\.]+)`)
	newRegex         = regexp.MustCompile(`^A\s+(?P<from>[\w\/\.-]+)$`)
	goneRegex        = regexp.MustCompile(`^D\s+(?P<from>[\w\/\.-]+)$`)
	modifiedRegex    = regexp.MustCompile(`^M\s+(?P<from>[\w\/\.-]+)$`)
	renamedRegex     = regexp.MustCompile(`^R\d{3}\s+(?P<from>[\w\/\.-]+)\s+(?P<to>[\w/\.-]+)$`)
	copyRegex        = regexp.MustCompile(`^C\d{3}\s+(?P<from>[\w\/\.-]+)\s+(?P<to>[\w/\.-]+)$`)
)

// Repository stores information about a Git repository
type Repository struct {
	mu  sync.Mutex
	dir string
	url string
}

// FetchOptions stores options on fetching updates
type FetchOptions struct {
}

// PullOptions stores options on pulling updates
type PullOptions struct {
	// The name of the remote to pull
	RemoteName string
	// The name of the reference to pull. If empty it uses master
	ReferenceName ReferenceName
}

// LogOptions stores options on Logging commands
type LogOptions struct {
	From Hash
}

// Remotes returns an iterator of the Repository's Remotes
func (r *Repository) Remotes() (RemoteIter, error) {
	if r == nil {
		return nil, errors.New("No repository")
	}
	log.Debugf("Getting Remotes for repository %v, stored in %v", r.url, r.dir)

	remCmd := exec.Command("git", "remote", "-v")
	remCmd.Dir = r.dir
	remOut, err := remCmd.Output()
	if err != nil {
		return nil, err
	}

	remotesRegex := regexp.MustCompile(`(\w+)\s+([\w\/:\-\.@]+)\s+\((fetch|push)\)`)

	remMap := make(map[string]*RemoteConfig)
	for _, match := range remotesRegex.FindAllStringSubmatch(string(remOut), -1) {
		rName := match[1]
		rURL := match[2]

		if _, ok := remMap[rName]; !ok {
			remMap[rName] = &RemoteConfig{
				Name: rName,
			}
		}

		rcfg := remMap[rName]
		rcfg.URLs = append(rcfg.URLs, rURL)
	}

	rems := make([]Remote, 0)
	for _, rem := range remMap {
		rems = append(rems, Remote{
			c: rem,
		})
	}

	return &sliceRemoteIter{pos: 0, series: rems}, nil
}

// Branches returns an iterator of the Repository's Branches
func (r *Repository) Branches() (ReferenceIter, error) {
	log.Debugf("Getting Branches for repository %v, stored in %v", r.url, r.dir)
	if r == nil {
		return nil, errors.New("No repository")
	}

	// Get the branches for the repository
	refCmd := exec.Command("git", "show-ref")

	refCmd.Dir = r.dir
	refOut, err := refCmd.Output()
	if err != nil {
		return nil, err
	}

	refs := make([]Reference, 0)
	for _, match := range refsRegex.FindAllStringSubmatch(string(refOut), -1) {
		refs = append(refs, Reference{
			t: HashReference,
			n: ReferenceName(match[2]),
			h: NewHash(match[1]),
		})
	}
	return &sliceRefIter{pos: 0, series: refs}, nil
}

// Log returns a CommitIter of the Repository's commits
func (r *Repository) Log(opts *LogOptions) (CommitIter, error) {
	r.mu.Lock()
	log.Debugf("Calling Log on %v from: %v", r.dir, opts.From)
	//Checkout the repo at the given hash
	if err := r.checkout(opts.From); err != nil {
		r.mu.Unlock()
		return nil, err
	}

	// Call git log on the repository, and parse output into commit objects
	commits, err := r.getCommits(opts.From)
	if err != nil {
		r.mu.Unlock()
		return nil, err
	}
	r.mu.Unlock()

	return &sliceCommitIter{
		series: commits,
		pos:    0,
	}, nil
}

// FetchContext calls 'git fetch'
func (r *Repository) FetchContext(ctx context.Context, o *FetchOptions) error {
	if r == nil {
		return errors.New("FetchContext was called with a nil repository")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fetchCmd := exec.Command("git", "fetch")
	fetchCmd.Dir = r.dir

	fetchOut, err := fetchCmd.CombinedOutput()
	if err != nil {
		log.Errorf("Hit an error fetching %v, in %v. \n\tError: %v.\n\tLog: %s", r.url, r.dir, err, fetchOut)
		return err
	}

	// If we get no output, we are up to date
	if len(fetchOut) == 0 {
		return ErrAlreadyUpToDate
	}

	return nil
}

// PullContext runs git pull with the given cancelation context
func (r *Repository) PullContext(ctx context.Context, o *PullOptions) error {
	if r == nil {
		return errors.New("PullContext was called with a nil repository")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	remoteName := "origin"
	if o.RemoteName != "" {
		remoteName = o.RemoteName
	}

	referenceName := "master"
	if o.ReferenceName != "" {
		referenceName = string(o.ReferenceName)
	}

	// Pull all remotes
	pullCmd := exec.Command("git", "pull", remoteName, referenceName)
	pullCmd.Dir = r.dir

	pullOut, err := pullCmd.Output()
	if err != nil {
		log.Errorf("Hit an error pulling %v, in %v. \n\tError: %v\n\t:Log: %v", r.url, r.dir, err, string(pullOut))
		return err
	}

	if strings.HasSuffix(string(pullOut), "Already up to date.\n") {
		return ErrAlreadyUpToDate
	}

	return nil
}

func (r *Repository) checkout(from Hash) error {
	checkoutCmd := exec.Command("git", "checkout", from.String(), "--force")
	checkoutCmd.Dir = r.dir

	checkoutOut, err := checkoutCmd.Output()
	if err != nil {
		log.Error(string(checkoutOut))
	}
	return err
}

func (r *Repository) getCommits(from Hash) ([]*Commit, error) {
	log.Debugf("Getting the commits for: %v stored in %v", r.url, r.dir)

	logCmdStr := fmt.Sprintf(`git rev-list %s --reverse | `, from.String()) + `
    while read sha1; do
        git show -s --format="Commit: %H%nAuthor: %an%nAuthor Email: %ae%nAuthor Date: %at%nCommitter: %cn%nCommitter Email: %ce%nCommitter Date: %ct%nSubject: %s" $sha1;
        git show -s --format="Body: %B" $sha1 | tr "\n" " " | tr "\r" " "; echo;
        echo "Files:"
        git show --format='' --name-status $sha1; echo;
        echo;
	done`
	logCmd := exec.Command("bash", "-c", logCmdStr)
	logCmd.Dir = r.dir

	logOut, err := logCmd.Output()
	if err != nil {
		return nil, err
	}
	logStr := string(logOut)

	commitMatches := commitItemRegExp.FindAllStringSubmatch(logStr, -1)
	commits := make([]*Commit, len(commitMatches))

	for idx, match := range commitMatches {
		// Extract commit data
		sha := match[1]
		author := match[2]
		aEmail := match[3]
		aDate := match[4]
		committer := match[5]
		cEmail := match[6]
		cDate := match[7]
		subject := match[8]
		body := match[9]
		changedFiles := match[10]

		atime, err := strconv.Atoi(aDate)
		if err != nil {
			log.Warnf("bad author time: %v", aDate)
			continue
		}
		aDateUnix := time.Unix(int64(atime), 0)

		ctime, err := strconv.Atoi(cDate)
		if err != nil {
			log.Warnf("bad commit time: %v", cDate)
			continue
		}
		cDateUnix := time.Unix(int64(ctime), 0)

		// Extract File Data
		if err = r.checkout(NewHash(sha)); err != nil {
			log.Errorf("Could not checkout: %v, Error: %v", sha, err)
		}

		fileStrings := strings.Split(changedFiles, "\n")

		// Can't use straight indexing into our array here as a single
		// line can contain multiple files (in the case of renames and copies)
		files := make([]File, 0)
		for _, mFile := range fileStrings {

			if is, p := isGone(mFile); is {
				log.Debugf("commit: %v file was delted %v", sha, p)
				nfiles, err := appendDelete(r.dir, p, files)
				if err != nil {
					return nil, err
				}
				files = nfiles
			} else if is, p := isNew(mFile); is {
				log.Debugf("commit: %v file is new %v", sha, p)
				nfiles, err := appendRegular(r.dir, p, files)
				if err != nil {
					return nil, err
				}
				files = nfiles
			} else if is, from, to := isCopy(mFile); is {
				log.Debugf("commit: %v file is copied from %v to %v", sha, from, to)
				log.Debug(mFile)

				nfiles, err := appendRegular(r.dir, from, files)
				if err != nil && !os.IsNotExist(err) {
					return nil, err
				}
				files = nfiles

				nfiles, err = appendRegular(r.dir, to, files)
				if err != nil && !os.IsNotExist(err) {
					return nil, err
				}
				files = nfiles
			} else if is, p := isModified(mFile); is {
				log.Debugf("commit: %v file is modified %v", sha, p)

				nfiles, err := appendRegular(r.dir, p, files)
				if err != nil {
					return nil, err
				}
				files = nfiles
			} else if is, from, to := isRenamed(mFile); is {
				log.Debugf("commit: %v file was renamed from: %v to: %v", sha, from, to)

				nfiles, err := appendDelete(r.dir, from, files)
				if err != nil {
					return nil, err
				}

				files = nfiles

				nfiles, err = appendRegular(r.dir, to, files)
				if err != nil {
					return nil, err
				}
				files = nfiles
			} else if s := strings.TrimSpace(mFile); s != "" {
				// In this case, we have hit something we shouldnt
				// this needs a major warning and a continue
				log.Warnf("commit: %v Unrecognized file state: %v", sha, mFile)
			}
		}

		commit := &Commit{
			Hash:      NewHash(sha),
			Author:    Signature{Name: author, Email: aEmail, When: aDateUnix},
			Committer: Signature{Name: committer, Email: cEmail, When: cDateUnix},
			Message:   fmt.Sprintf("%s\n%s", subject, body),
			files:     files,
		}
		commits[idx] = commit
	}

	if len(commits) != len(commitMatches) {
		log.Warnf("Our regex found: %v commits, but %v were added to our array", len(commits), len(commitMatches))
	}

	debug.FreeOSMemory()

	log.Debugf("Returning %v commits for %v checked out in %v", len(commits), r.url, r.dir)
	return commits, nil
}

func isNew(mFile string) (bool, string) {
	if !newRegex.MatchString(mFile) {
		return false, ""
	}
	m := newRegex.FindAllStringSubmatch(mFile, -1)
	return true, m[0][1]
}

func isGone(mFile string) (bool, string) {
	if !goneRegex.MatchString(mFile) {
		return false, ""
	}
	m := goneRegex.FindAllStringSubmatch(mFile, -1)
	return true, m[0][1]
}

func isModified(mFile string) (bool, string) {
	if !modifiedRegex.MatchString(mFile) {
		return false, ""
	}
	m := modifiedRegex.FindAllStringSubmatch(mFile, -1)
	return true, m[0][1]
}

func isRenamed(mFile string) (bool, string, string) {
	if !renamedRegex.MatchString(mFile) {
		return false, "", ""
	}
	m := renamedRegex.FindStringSubmatch(mFile)

	result := make(map[string]string)
	for i, name := range renamedRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = m[i]
		}
	}

	return true, result["from"], result["to"]
}

func isCopy(mFile string) (bool, string, string) {
	if !copyRegex.MatchString(mFile) {
		return false, "", ""
	}
	m := copyRegex.FindStringSubmatch(mFile)

	result := make(map[string]string)
	for i, name := range copyRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = m[i]
		}
	}

	return true, result["from"], result["to"]
}

func appendRegular(dir, filename string, files []File) ([]File, error) {
	fullPath := path.Join(dir, filename)

	fl, err := os.Lstat(fullPath)
	if err != nil {
		log.Errorf("Error: %v", err)
		return files, err
	}
	if fl.Mode()&os.ModeSymlink != 0 {
		// In this case, we have a symlink
		// SKIP IT
		return files, nil
	}

	fi, err := os.Stat(fullPath)
	if err != nil {
		log.Errorf("Error: %v", err)
		return files, err
	}
	if fi.IsDir() {
		return files, nil
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Errorf("Error: %v", err)
		return files, err
	}

	return append(files, File{
		Name:     filename,
		Size:     fi.Size(),
		contents: string(contents),
	}), nil
}

func appendDelete(dir, filename string, files []File) ([]File, error) {
	return append(files, File{
		Name:     filename,
		Size:     0,
		contents: "",
	}), nil
}
