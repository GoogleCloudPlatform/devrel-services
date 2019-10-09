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
	"io"
	"testing"
)

func TestFileContentsReturnsContents(t *testing.T) {
	file := File{contents: "Hello World"}
	if c, err := file.Contents(); c != "Hello World" || err != nil {
		t.Errorf("Expected %v, %v. Got %v, %v", "Hello World", "nil", c, err)
	}

	file = File{contents: ""}
	if c, err := file.Contents(); c != "" || err != nil {
		t.Errorf("Expected %v, %v. Got %v, %v", "", "nil", c, err)
	}
}

func TestSliceFileIterForEachReturnsInProperOrder(t *testing.T) {
	files := []File{
		File{},
		File{},
		File{},
	}

	iter := &sliceFileIter{
		pos:    0,
		series: files,
	}

	pos := 0
	iter.ForEach(func(c *File) error {
		if files[pos] != *c {
			t.Errorf("The order returned does not match for idex: %v", pos)
		}
		pos++
		return nil
	})
}

func TestSliceFileIterForEachReturnsEarlyOnError(t *testing.T) {
	files := []File{
		File{},
		File{},
		File{},
	}

	iter := &sliceFileIter{
		pos:    0,
		series: files,
	}

	expErr := errors.New("Short circuit")
	pos := 0
	err := iter.ForEach(func(c *File) error {
		if pos == 1 {
			return expErr
		}
		pos++
		return nil
	})

	if pos != 1 || err != expErr {
		t.Errorf("Expected to short circuit at position 1. Got: %v %v", pos, err)
	}
}

func TestSliceFileIterForEachReturnsEarlyOnErrorSignal(t *testing.T) {
	files := []File{
		File{},
		File{},
		File{},
	}

	iter := &sliceFileIter{
		pos:    0,
		series: files,
	}

	pos := 0
	err := iter.ForEach(func(c *File) error {
		if pos == 1 {
			return ErrStop
		}
		pos++
		return nil
	})

	if pos != 1 || err != nil {
		t.Errorf("Expected to short circuit at position 1. Got: %v %v", pos, err)
	}
}

func TestSliceFileIterNextReturnsInProperOrder(t *testing.T) {
	files := []File{
		File{},
		File{},
		File{},
	}

	iter := &sliceFileIter{
		pos:    0,
		series: files,
	}

	for idx := 0; idx < len(files); idx++ {
		if c, err := iter.Next(); *c != files[idx] || err != nil {
			t.Errorf("Expected %v, %v, got %v, %v", files[idx], "nil", c, err)
		}
	}

	c, err := iter.Next()
	if c != nil || err != io.EOF {
		t.Errorf("Next at the end of the iter should return EOF. It returned %v, %v", c, err)
	}
}

func TestSliceFileIterCloseEndsEarly(t *testing.T) {
	files := []File{
		File{},
		File{},
		File{},
	}

	iter := &sliceFileIter{
		pos:    0,
		series: files,
	}

	c0, err := iter.Next()
	if *c0 != files[0] || err != nil {
		t.Errorf("sliceFileIter returned unexpected values %v, %v", c0, err)
	}

	iter.Close()
	c1, err := iter.Next()
	if c1 != nil || err != io.EOF {
		t.Errorf("After closing, sliceFileIter returned unexpected values. Expected nil, io.EOF, got %v, %v", c1, err)
	}
}
