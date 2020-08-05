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

func TestFilesReturnsAFileIter(t *testing.T) {
	commit := &Commit{}
	iter, err := commit.Files()
	if iter == nil || err != nil {
		t.Errorf("Expected an iter and no error. Got %v %v", iter, err)
	}
}

func TestSliceCommitIterForEachReturnsInProperOrder(t *testing.T) {
	commits := []*Commit{
		&Commit{},
		&Commit{},
		&Commit{},
	}

	iter := &sliceCommitIter{
		pos:    0,
		series: commits,
	}

	pos := 0
	iter.ForEach(func(c *Commit) error {
		if commits[pos] != c {
			t.Errorf("The order returned does not match for idex: %v", pos)
		}
		pos++
		return nil
	})
}

func TestSliceCommitIterNextReturnsInProperOrder(t *testing.T) {
	commits := []*Commit{
		&Commit{},
		&Commit{},
		&Commit{},
	}

	iter := &sliceCommitIter{
		pos:    0,
		series: commits,
	}

	for idx := 0; idx < len(commits); idx++ {
		if c, err := iter.Next(); c != commits[idx] || err != nil {
			t.Errorf("Expected %v, %v, got %v, %v", commits[idx], "nil", c, err)
		}
	}

	c, err := iter.Next()
	if c != nil || err != io.EOF {
		t.Errorf("Next at the end of the iter should return EOF. It returned %v, %v", c, err)
	}
}

func TestSliceCommitIterForEachReturnsEarlyOnError(t *testing.T) {
	commits := []*Commit{
		&Commit{},
		&Commit{},
		&Commit{},
	}

	iter := &sliceCommitIter{
		pos:    0,
		series: commits,
	}

	expErr := errors.New("Short circuit")
	pos := 0
	err := iter.ForEach(func(c *Commit) error {
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

func TestSliceCommitIterForEachReturnsEarlyOnErrorSignal(t *testing.T) {
	commits := []*Commit{
		&Commit{},
		&Commit{},
		&Commit{},
	}

	iter := &sliceCommitIter{
		pos:    0,
		series: commits,
	}

	pos := 0
	err := iter.ForEach(func(c *Commit) error {
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

func TestSliceCommiterCloseEndsEarly(t *testing.T) {
	commits := []*Commit{
		&Commit{},
		&Commit{},
		&Commit{},
	}

	iter := &sliceCommitIter{
		pos:    0,
		series: commits,
	}

	c0, err := iter.Next()
	if c0 != commits[0] || err != nil {
		t.Errorf("sliceCommitIter returned unexpected values %v, %v", c0, err)
	}

	iter.Close()
	c1, err := iter.Next()
	if c1 != nil || err != io.EOF {
		t.Errorf("After closing, sliceCommitIter returned unexpected values. Expected nil, io.EOF, got %v, %v", c1, err)
	}
}
