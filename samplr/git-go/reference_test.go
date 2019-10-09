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

func TestCreatesProperReference(t *testing.T) {
	name := "Reference"
	hash := NewHash("f3f4e32b94c97ffd6b2e6f020d96931a9fe2c3f6")
	rType := HashReference

	ref := Reference{
		t: rType,
		h: hash,
		n: ReferenceName(name),
	}

	if refName := ref.Name().String(); refName != name {
		t.Errorf("Reference Name. Expected %v, Got %v", name, refName)
	}
	if refHash := ref.Hash(); refHash != hash {
		t.Errorf("Reference Hash. Expected %v, Got %v", name, refHash)
	}
}

func TestSliceRefIterForEachReturnsInProperOrder(t *testing.T) {
	refs := []Reference{
		Reference{},
		Reference{},
		Reference{},
	}

	iter := &sliceRefIter{
		pos:    0,
		series: refs,
	}

	pos := 0
	iter.ForEach(func(c *Reference) error {
		if refs[pos] != *c {
			t.Errorf("The order returned does not match for idex: %v", pos)
		}
		pos++
		return nil
	})
}

func TestSliceRefIterNextReturnsInProperOrder(t *testing.T) {
	refs := []Reference{
		Reference{},
		Reference{},
		Reference{},
	}

	iter := &sliceRefIter{
		pos:    0,
		series: refs,
	}

	for idx := 0; idx < len(refs); idx++ {
		if c, err := iter.Next(); *c != refs[idx] || err != nil {
			t.Errorf("Expected %v, %v, got %v, %v", refs[idx], "nil", c, err)
		}
	}

	c, err := iter.Next()
	if c != nil || err != io.EOF {
		t.Errorf("Next at the end of the iter should return EOF. It returned %v, %v", c, err)
	}
}

func TestSliceRefIterForEachReturnsEarlyOnError(t *testing.T) {
	refs := []Reference{
		Reference{},
		Reference{},
		Reference{},
	}

	iter := &sliceRefIter{
		pos:    0,
		series: refs,
	}

	expErr := errors.New("Short circuit")
	pos := 0
	err := iter.ForEach(func(c *Reference) error {
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

func TestSliceRefIterForEachReturnsEarlyOnErrorSignal(t *testing.T) {
	refs := []Reference{
		Reference{},
		Reference{},
		Reference{},
	}

	iter := &sliceRefIter{
		pos:    0,
		series: refs,
	}

	pos := 0
	err := iter.ForEach(func(c *Reference) error {
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

func TestSliceReferCloseEndsEarly(t *testing.T) {
	refs := []Reference{
		Reference{},
		Reference{},
		Reference{},
	}

	iter := &sliceRefIter{
		pos:    0,
		series: refs,
	}

	c0, err := iter.Next()
	if *c0 != refs[0] || err != nil {
		t.Errorf("sliceRefIter returned unexpected values %v, %v", c0, err)
	}

	iter.Close()
	c1, err := iter.Next()
	if c1 != nil || err != io.EOF {
		t.Errorf("After closing, sliceRefIter returned unexpected values. Expected nil, io.EOF, got %v, %v", c1, err)
	}
}
