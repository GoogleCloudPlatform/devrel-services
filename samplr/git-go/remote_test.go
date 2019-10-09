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

func TestSliceRemoteIterForEachReturnsInProperOrder(t *testing.T) {
	rems := []Remote{
		Remote{},
		Remote{},
		Remote{},
	}

	iter := &sliceRemoteIter{
		pos:    0,
		series: rems,
	}

	pos := 0
	iter.ForEach(func(c *Remote) error {
		if rems[pos] != *c {
			t.Errorf("The order returned does not match for idex: %v", pos)
		}
		pos++
		return nil
	})
}

func TestSliceRemoteIterNextReturnsInProperOrder(t *testing.T) {
	rems := []Remote{
		Remote{},
		Remote{},
		Remote{},
	}

	iter := &sliceRemoteIter{
		pos:    0,
		series: rems,
	}

	for idx := 0; idx < len(rems); idx++ {
		if c, err := iter.Next(); *c != rems[idx] || err != nil {
			t.Errorf("Expected %v, %v, got %v, %v", rems[idx], "nil", c, err)
		}
	}

	c, err := iter.Next()
	if c != nil || err != io.EOF {
		t.Errorf("Next at the end of the iter should return EOF. It returned %v, %v", c, err)
	}
}

func TestSliceRemoteIterForEachReturnsEarlyOnError(t *testing.T) {
	rems := []Remote{
		Remote{},
		Remote{},
		Remote{},
	}

	iter := &sliceRemoteIter{
		pos:    0,
		series: rems,
	}

	expErr := errors.New("Short circuit")
	pos := 0
	err := iter.ForEach(func(c *Remote) error {
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

func TestSliceRemoteIterForEachReturnsEarlyOnErrorSignal(t *testing.T) {
	rems := []Remote{
		Remote{},
		Remote{},
		Remote{},
	}

	iter := &sliceRemoteIter{
		pos:    0,
		series: rems,
	}

	pos := 0
	err := iter.ForEach(func(c *Remote) error {
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

func TestSliceRemoteIterCloseEndsEarly(t *testing.T) {
	rems := []Remote{
		Remote{},
		Remote{},
		Remote{},
	}

	iter := &sliceRemoteIter{
		pos:    0,
		series: rems,
	}

	c0, err := iter.Next()
	if *c0 != rems[0] || err != nil {
		t.Errorf("sliceRemoteIter returned unexpected values %v, %v", c0, err)
	}

	iter.Close()
	c1, err := iter.Next()
	if c1 != nil || err != io.EOF {
		t.Errorf("After closing, sliceRemoteIter returned unexpected values. Expected nil, io.EOF, got %v, %v", c1, err)
	}
}
