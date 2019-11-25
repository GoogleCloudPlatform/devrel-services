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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
)

func TestClones(t *testing.T) {
	cases := []struct {
		Name string
		URL  string
	}{
		{
			Name: "java-docs-samples",
			URL:  "https://github.com/GoogleCloudPlatform/java-docs-samples",
		},
		{
			Name: "cloud-code-intellij",
			URL:  "https://github.com/GoogleCloudPlatform/cloud-code-intellij",
		},
	}

	for _, c := range cases {
		dir, err := ioutil.TempDir("", "git-go-test")
		if err != nil {
			t.Errorf("Test: %v, Got Err making temp dir: %v", c.Name, err)
			continue
		}
		r, err := PlainClone(dir, true, &CloneOptions{URL: c.URL})

		if err != nil {
			t.Errorf("Test: %v, Got Err cloning: %v", c.Name, err)
			continue
		}

		var ref Hash

		riter, err := r.Branches()
		if err != nil {
			t.Errorf("Test %v, Got Err getting remotes %v", c.Name, err)
			continue
		}

		riter.ForEach(func(referene *Reference) error {
			if referene.Name() == Master {
				ref = referene.Hash()
			}
			return nil
		})

		if ref.IsZero() {
			t.Errorf("Test %v, Got Zero Hash when finding master branch", c.Name)
			continue
		}

		iter, err := r.Log(&LogOptions{
			From: ref,
		})

		if err != nil {
			t.Errorf("Test %v, Got error creating log: %v", c.Name, err)
			continue
		}

		nCommits := 0
		iter.ForEach(func(c *Commit) error {
			nCommits = nCommits + 1
			return nil
		})
		if nCommits < 1 {
			t.Errorf("Test %v, nCommits should be more than 0. It is: %v", c.Name, nCommits)
		}
	}
}

func TestFetch(t *testing.T) {
	odir, err := ioutil.TempDir("", "origin")
	if err != nil {
		t.Errorf("TestFetch failed creating origin dir: %v", err)
	}
	defer os.RemoveAll(odir)

	initCmd := exec.Command("git", "init", "--bare")
	initCmd.Dir = odir
	_, err = initCmd.Output()
	if err != nil {
		t.Errorf("TestFetch failed init-ing git repository: %v", err)
	}

	c1dir, err := ioutil.TempDir("", "clone1")
	if err != nil {
		t.Errorf("TestFetch failed creating clone1 dir: %v", err)
	}
	defer os.RemoveAll(c1dir)

	c2dir, err := ioutil.TempDir("", "clone2")
	if err != nil {
		t.Errorf("TestFetch failed creating clone1 dir: %v", err)
	}
	defer os.RemoveAll(c2dir)

	repo, err := PlainClone(c1dir, true, &CloneOptions{URL: odir})

	if err != nil {
		t.Errorf("TestFetch, Got Err cloning: %v", err)
	}

	// Ensure that fetch after a fresh clone returns no error
	err = repo.FetchContext(context.Background(), &FetchOptions{})
	if err != NoErrAlreadyUpToDate {
		t.Errorf("TestFetch, Got an error on first fetch: %v", err)
	}

	// Standard Clone into c2dir
	c2CloneCmd := exec.Command("git", "clone", odir, c2dir)
	_, err = c2CloneCmd.Output()
	if err != nil {
		t.Errorf("TestFetch, got an error cloning 2: %v", err)
	}

	// Add a file and commit/clone it from c2Clone
	lcmd := exec.Command("touch", "foo")
	lcmd.Dir = c2dir
	err = lcmd.Run()
	if err != nil {
		t.Errorf("TestFetch error adding commit: %v", err)
	}

	lcmd = exec.Command("git", "add", "foo")
	lcmd.Dir = c2dir
	err = lcmd.Run()
	if err != nil {
		t.Errorf("TestFetch error adding commit: %v", err)
	}

	lcmd = exec.Command("git", "commit", "-m", "\"foo\"")
	lcmd.Dir = c2dir
	err = lcmd.Run()
	if err != nil {
		t.Errorf("TestFetch error adding commit: %v", err)
	}

	lcmd = exec.Command("git", "push")
	lcmd.Dir = c2dir
	err = lcmd.Run()
	if err != nil {
		t.Errorf("TestFetch error adding commit: %v", err)
	}

	// Call fetch on r1 and assume err is nil
	err = repo.FetchContext(context.Background(), &FetchOptions{})
	if err != nil {
		t.Errorf("TestFetch: got an error on second fetch: %v", err)
	}

	// Now call Pull on repo
	err = repo.PullContext(context.Background(), &PullOptions{})
	if err != nil {
		t.Errorf("TestFetch: got an error on pulling: %v", err)
	}

	// Since we have pulled the second pull should return NoErr
	err = repo.PullContext(context.Background(), &PullOptions{})
	if err != NoErrAlreadyUpToDate {
		t.Errorf("TestFetch: got an error on second pull: %v", err)
	}

	// Finally check that c1dir has foo in it
	if _, err := os.Stat(path.Join(c1dir, "foo")); os.IsNotExist(err) {
		t.Errorf(("TestFetch: c1dir/foo does not exist"))
	}
}
