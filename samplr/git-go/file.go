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

import "io"

// FilesIter is a generic closable interface for iterating over files.
type FilesIter interface {
	Next() (*File, error)
	ForEach(func(*File) error) error
	Close()
}

// File represents a git file
type File struct {
	Name     string
	Size     int64
	contents string
}

// Contents returns the contents of a file
func (f File) Contents() (string, error) {
	return f.contents, nil
}

type sliceFileIter struct {
	pos    int
	series []File
}

func (iter *sliceFileIter) Next() (*File, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}

	obj := iter.series[iter.pos]
	iter.pos++
	return &obj, nil
}

func (iter *sliceFileIter) ForEach(fn func(*File) error) error {
	defer iter.Close()
	for _, r := range iter.series {
		if err := fn(&r); err != nil {
			if err == ErrStop {
				return nil
			}

			return err
		}
	}
	return nil
}

func (iter *sliceFileIter) Close() {
	iter.pos = len(iter.series)
}
