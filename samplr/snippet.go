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

package samplr

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"

	// git "gopkg.in/src-d/go-git.v4"
	git "github.com/GoogleCloudPlatform/devrel-services/git-go"

	enry "gopkg.in/src-d/enry.v1"
)

var startReg = regexp.MustCompile(`\[START ([\w-_]+)\]`)

// There is DevRel tooling that puts region tags in in pydoc comments for use in
// generating reference docs. We need to ignore these usages.
var startIgnoreReg = regexp.MustCompile(`:start-after:`)
var endIgnoreReg = regexp.MustCompile(`:end-before:`)

const (
	startRegTemplate = `\[START %v\]`
	endRegTemplate   = `\[END %v\]`
	linesFormat      = `L%v-L%v`
)

var fileWhitelist = []*regexp.Regexp{
	regexp.MustCompile("^.+\\.c$"),           // c
	regexp.MustCompile("^.+\\.cpp$"),         // cpp
	regexp.MustCompile("^.+\\.cc$"),          // cpp
	regexp.MustCompile("^.+\\.cs$"),          // csharp
	regexp.MustCompile("^(?i)(dockerfile)$"), // dockerfile
	regexp.MustCompile("^.+\\.go$"),          // go
	regexp.MustCompile("^.+\\.gs$"),          // apps_script
	regexp.MustCompile("^.+\\.html$"),        // index.html, etc.
	regexp.MustCompile("^.+\\.jade$"),        // Node.js jade template files
	regexp.MustCompile("^.+\\.java$"),        // java
	regexp.MustCompile("^.+\\.js$"),          // javascript, node.js
	regexp.MustCompile("^.+\\.json$"),        // package.json, etc.
	regexp.MustCompile("^.+\\.(kt|kts)$"),    // kotlin
	regexp.MustCompile("^.+\\.m$"),           // ios_objc
	regexp.MustCompile("^.+\\.php$"),         // php
	regexp.MustCompile("^.+\\.pug$"),         // Node.js pug template files
	regexp.MustCompile("^.+\\.py$"),          // python
	regexp.MustCompile("^.+\\.rb$"),          // ruby
	regexp.MustCompile("^.+\\.ru$"),          // ruby
	regexp.MustCompile("^.+\\.swift$"),       // swift
	regexp.MustCompile("^.+\\.sh$"),          // bash
	regexp.MustCompile("^.+\\.xml$"),         // pmx.xml etc.
	regexp.MustCompile("^.+\\.yaml$"),        // app.yaml, etc.
}

// Snippet represents a snippet of code
type Snippet struct {
	Name     string
	Language string
	Versions []SnippetVersion
	Primary  SnippetVersion
}

// SnippetVersion represents a snippet at a particular commit in a repository
type SnippetVersion struct {
	Name    string
	File    *File
	Lines   []string
	Content string
	Meta    SnippetVersionMeta
}

type SnippetVersionMeta struct {
	Title       string
	Description string
	Usage       string
	ApiVersion  string
}

// File represents a file at a git commit
type File struct {
	FilePath  string
	GitCommit *GitCommit
	Size      int64
}

// GitCommit represents a commit in git
type GitCommit struct {
	Body           string
	Subject        string
	AuthorEmail    string
	AuthoredTime   time.Time
	CommitterEmail string
	CommittedTime  time.Time
	Hash           string
	Name           string
}

func CalculateSnippets(o, r string, iter git.CommitIter) ([]*Snippet, error) {
	log.Debugf("Calculating snippets for: %v/%v", o, r)
	snippets := make(map[string]*Snippet, 0)
	seenSnippets := make(map[string]map[string]bool, 0)

	err := iter.ForEach(func(co *git.Commit) error {
		commit := co
		filesIter, err := commit.Files()
		if err != nil {
			log.Warnf("Repo: %v/%v. Got error in commit.Files() for commit %v: %v", o, r, commit.Hash.String(), err)
			return err
		}

		cmt := GitCommit{
			Body:           commit.Message,
			Subject:        strings.Split(commit.Message, "\n")[0],
			AuthorEmail:    commit.Author.Email,
			AuthoredTime:   commit.Author.When,
			CommitterEmail: commit.Committer.Email,
			CommittedTime:  commit.Committer.When,
			Hash:           commit.Hash.String(),
			Name:           fmt.Sprintf("owners/%s/repositories/%s/gitCommits/%s", o, r, commit.Hash.String()),
		}

		snippetVersionsInThisCommit := make(map[*File]map[string]SnippetVersion, 0)
		deletedFilesInThisCommit := make(map[string]bool, 0)

		err = filesIter.ForEach(func(file *git.File) error {
			log.Debugf("Processing commit: %v, file: %v", cmt.Hash, file.Name)
			if !isValidFile(file) {
				log.Debugf("Processing commit: %v, Invalid file: %v", cmt.Hash, file.Name)
				return nil
			}

			content, err := file.Contents()
			if err != nil {
				log.Errorf("Processing commit: %v. Hit an error getting file contents for file: %v, %v", cmt.Hash, file.Name, err)
				return err
			}

			// File is deleted.
			if content == "" && file.Size == 0 {
				log.Debugf("Processing commit: %v Adding: %v to deletedFilesInThisCommit", cmt.Hash, file.Name)
				deletedFilesInThisCommit[file.Name] = true
				return nil
			}

			language := enry.GetLanguage(file.Name, []byte(content))
			if language == "C#" {
				language = "CSHARP"
			} else if language == "C++" {
				language = "CPP"
			}
			language = strings.ToUpper(language)

			fle := File{
				FilePath:  file.Name,
				GitCommit: &cmt,
				Size:      file.Size,
			}

			// The %%s is an escaped "%s". So that when  fmt.Sprintf() formats
			// the string, there will still be an "%s". This allows the
			// extractSnippetVersionFromFile function to fill in the tag name
			snpNameFmt := fmt.Sprintf("owners/%s/repositories/%s/snippets/%%s/languages/%s", o, r, language)

			snpVersions, err := extractSnippetVersionsFromFile(content, snpNameFmt)
			if err != nil {
				return err
			}

			for _, v := range snpVersions {
				if _, ok := snippets[v.Name]; !ok {
					snippets[v.Name] = &Snippet{
						Name:     v.Name,
						Language: language,
						Versions: make([]SnippetVersion, 0),
					}
				}
			}

			snippetVersionsInThisCommit[&fle] = snpVersions
			return nil
		})

		snippets = processDeletedFiles(&cmt, snippets, deletedFilesInThisCommit, seenSnippets)

		processPreviouslySeenSnippets(&cmt, snippets, snippetVersionsInThisCommit, seenSnippets)

		processSnippetVersionsInThisCommit(&cmt, snippets, snippetVersionsInThisCommit, seenSnippets)

		return err
	})

	retval := make([]*Snippet, len(snippets))
	idx := 0
	for _, snp := range snippets {
		retval[idx] = snp
		idx++
	}

	log.Debugf("For repository: %v/%v, returning %v snippets", o, r, len(retval))

	return retval, err
}

func extractSnippetVersionsFromFile(content string, nfmt string) (map[string]SnippetVersion, error) {
	versionTags := make(map[string]SnippetVersion, 0)

	sampleMeta, err := parseSampleMetadata(content)
	if err != nil {
		return versionTags, err
	}

	for _, tag := range detectRegionTags(content) {
		scn := bufio.NewScanner(strings.NewReader(content))

		reStart, err := regexp.Compile(fmt.Sprintf(startRegTemplate, tag))
		if err != nil {
			return nil, err
		}
		reEnd, err := regexp.Compile(fmt.Sprintf(endRegTemplate, tag))
		if err != nil {
			return nil, err
		}

		lnc := 1
		type startEnd struct {
			Start int
			End   int
		}
		lnpairs := make([]startEnd, 0)
		combinedContent := strings.Builder{}
		lnp := startEnd{
			Start: -1,
			End:   -1,
		}
		for scn.Scan() {
			txt := scn.Text()
			if reStart.MatchString(txt) && !startIgnoreReg.MatchString(txt) {
				lnp.Start = lnc
			} else if lnp.Start != -1 && reEnd.MatchString(txt) && !endIgnoreReg.MatchString(txt) {
				lnp.End = lnc
				combinedContent.WriteString(txt)
				combinedContent.WriteString("\n")
				lnpairs = append(lnpairs, lnp)
				lnp = startEnd{
					Start: -1,
					End:   -1,
				}
			}

			if lnp.Start != -1 {
				combinedContent.WriteString(txt)
				combinedContent.WriteString("\n")
			}
			lnc++
		}

		lines := make([]string, 0)
		for _, lnp := range lnpairs {
			if lnp.Start != -1 && lnp.End != -1 {
				lines = append(lines, fmt.Sprintf(linesFormat, lnp.Start, lnp.End))
			}
		}

		if len(lines) > 0 {
			meta := SnippetVersionMeta{}
			if sampleMeta != nil {
				meta = SnippetVersionMeta{
					Title:       sampleMeta.Meta.Title,
					Description: sampleMeta.Meta.Description,
					Usage:       sampleMeta.Meta.Usage,
					ApiVersion:  sampleMeta.Meta.ApiVersion,
				}
				for _, smeta := range sampleMeta.Meta.Snippets {
					if tag == smeta.RegionTag {
						if smeta.Description != "" {
							meta.Description = smeta.Description
						}
						if smeta.Usage != "" {
							meta.Usage = smeta.Usage
						}
						break
					}
				}
			}
			snpVersion := SnippetVersion{
				Name:    fmt.Sprintf(nfmt, tag),
				File:    nil,
				Lines:   lines,
				Content: strings.TrimSpace(combinedContent.String()),
				Meta:    meta,
			}

			versionTags[tag] = snpVersion
		}
	}
	return versionTags, nil
}

func detectRegionTags(content string) []string {
	tagNames := make([]string, 0)
	for _, mtc := range startReg.FindAllStringSubmatch(content, -1) {
		if len(mtc) > 0 {
			tagNames = append(tagNames, mtc[1])
		}
	}
	return tagNames
}

func isValidFile(file *git.File) bool {
	for _, pattern := range fileWhitelist {
		if pattern.MatchString(file.Name) {
			return true
		}
	}
	return false
}

func strEq(a, b []string) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func processDeletedFiles(cmt *GitCommit,
	snippets map[string]*Snippet,
	deletedFilesInThisCommit map[string]bool,
	seenSnippets map[string]map[string]bool) map[string]*Snippet {

	log.Debugf("Processing commit: %v, Number of files deleted: %v", cmt.Hash, len(deletedFilesInThisCommit))
	// Add all delete records
	for delFile := range deletedFilesInThisCommit {
		log.Debugf("Processing commit: %v. Ranging over deletedFilesInThisCommit", cmt.Hash)
		for seenFile, seenSnps := range seenSnippets {
			if delFile != seenFile {
				continue
			}
			log.Debugf("Processing commit: %v. File %v was deleted", cmt.Hash, delFile)
			for seenSnippet, seen := range seenSnps {
				if !seen {
					continue
				}
				log.Debugf("Processing commit: %v File: %v was deleted. Adding a delete record for Snippet: %v", cmt.Hash, delFile, seenSnippet)

				if _, ok := snippets[seenSnippet]; !ok {
					log.Warnf("Processing commit: %v Processing Deletes. Snippet %v was seen, but is not in our snippets collection", cmt.Hash, seenSnippet)
					continue
				}

				seenSnps[seenSnippet] = false
				// Insert an "empty" snippet version for this
				nFile := File{
					FilePath:  delFile,
					GitCommit: cmt,
					Size:      0,
				}

				snippets[seenSnippet].Versions = append(snippets[seenSnippet].Versions, SnippetVersion{
					Name:    fmt.Sprintf("%v/%v", seenSnippet, len(snippets[seenSnippet].Versions)),
					File:    &nFile,
					Lines:   make([]string, 0),
					Content: "",
				})

				snippets[seenSnippet].Primary = SnippetVersion{
					Name:    snippets[seenSnippet].Versions[len(snippets[seenSnippet].Versions)-1].Name,
					File:    snippets[seenSnippet].Versions[len(snippets[seenSnippet].Versions)-1].File,
					Lines:   make([]string, 0),
					Content: "",
				}
			}
		}
	}

	return snippets
}

func processPreviouslySeenSnippets(cmt *GitCommit, snippets map[string]*Snippet,
	snippetVersionsInThisCommit map[*File]map[string]SnippetVersion,
	seenSnippets map[string]map[string]bool) {

	log.Debugf("Processing commit: %v, looking through previously seen snippets.", cmt.Hash)
	for seenFile, seenSnps := range seenSnippets {
		for seenSnippet, seen := range seenSnps {
			if seen {
				// We have seen this snippet.... check
				// if in this commit, the snippet exists
				// for this file
				found := false
				var foundFile *File = nil
				for fle, snippetVersions := range snippetVersionsInThisCommit {
					if found || fle.FilePath != seenFile {
						continue
					}

					foundFile = fle
					for _, version := range snippetVersions {
						if version.Name != seenSnippet {
							continue
						}
						found = true
					}
				}

				// If we found it, or if the file it was in was not changed in this commit,
				// then this snippet is unmodified
				if found || foundFile == nil {
					continue
				}

				log.Debugf("Processing commit: %v. We did not find snippet: %v. Adding a delete record", cmt.Hash, seenSnippet)

				seenSnps[seenSnippet] = false
				// Insert an "empty" snippet version for this
				nFile := File{
					FilePath:  foundFile.FilePath,
					GitCommit: cmt,
					Size:      foundFile.Size,
				}

				snippets[seenSnippet].Versions = append(snippets[seenSnippet].Versions, SnippetVersion{
					Name:    fmt.Sprintf("%v/%v", seenSnippet, len(snippets[seenSnippet].Versions)),
					File:    &nFile,
					Lines:   make([]string, 0),
					Content: "",
				})

				snippets[seenSnippet].Primary = SnippetVersion{
					Name:    snippets[seenSnippet].Versions[len(snippets[seenSnippet].Versions)-1].Name,
					File:    snippets[seenSnippet].Versions[len(snippets[seenSnippet].Versions)-1].File,
					Lines:   make([]string, 0),
					Content: "",
				}
			}
		}
	}
}

func processSnippetVersionsInThisCommit(cmt *GitCommit, snippets map[string]*Snippet,
	snippetVersionsInThisCommit map[*File]map[string]SnippetVersion,
	seenSnippets map[string]map[string]bool) {

	log.Debugf("Processing commit: %v. Adding snippetVersionsInThisCommit", cmt.Hash)
	// At this point........ we have updated our seenSnippets with the deletes
	// as well as the
	for fle, snippetVersions := range snippetVersionsInThisCommit {
		if _, ok := seenSnippets[fle.FilePath]; !ok {
			seenSnippets[fle.FilePath] = make(map[string]bool, 0)
		}

		for _, snippetVersion := range snippetVersions {
			snippetVersion.File = fle

			snippet := snippets[snippetVersion.Name]
			snippetVersion.Name = fmt.Sprintf("%v/%v", snippetVersion.Name, len(snippet.Versions))

			// At this point, if our current snippet version has the same
			// line numbers and content as the previous one.... we can
			// just continue. The file changed, but our snippet version
			// did not.
			if snippetsEquivalent(snippet.Primary, snippetVersion) {
				continue
			}

			log.Debugf("Processing commit: %v. Adding snippet version for snippet: %v", cmt.Hash, snippet.Name)

			snippet.Versions = append(snippet.Versions, snippetVersion)
			snippet.Primary = snippetVersion

			seenSnippets[fle.FilePath][snippet.Name] = true
		}
	}
}

func snippetsEquivalent(a SnippetVersion, b SnippetVersion) bool {
	// Check concrete values first
	if a.Content != b.Content {
		return false
	}
	// Why doesnt golang not have the XOR operator????
	if (a.File == nil || b.File == nil) && !(a.File == nil && b.File == nil) {
		// If one is nil and the other isnt they are not equal
		return false
	}
	//If they are both nil, they are equivalent
	if a.File == nil && b.File == nil {
		return true
	}
	// Last case, check the file paths are the same
	return a.File.FilePath == b.File.FilePath && strEq(a.Lines, b.Lines)
}
