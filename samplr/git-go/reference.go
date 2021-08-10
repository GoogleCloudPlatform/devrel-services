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
	"fmt"
	"io"
)

// ReferenceType reference type's
type ReferenceType int8

const (
	//InvalidReference is an Invalid Reference
	InvalidReference ReferenceType = 0
	// HashReference is a reference to a hash
	HashReference ReferenceType = 1
	// SymbolicReference is a symbolic reference
	SymbolicReference ReferenceType = 2
)

const (
	// HEAD is the name of the HEAD reference
	HEAD ReferenceName = "HEAD"
	// Master is the name of the Master reference
	Master ReferenceName = "refs/heads/master"
	// OriginMaster is the name of origin master
	OriginMaster ReferenceName = "refs/remotes/origin/master"
)

// ReferenceName reference name's
type ReferenceName string

// FullyQualifiedReferenceName takes a simple branch name and returns
// a ReferenceName appropriate for it. e.g. "main" => "refs/heads/main"
func FullyQualifiedReferenceName(n string) ReferenceName {
	return ReferenceName(fmt.Sprintf("refs/heads/%v", n))
}

// Reference represents a Git Reference
type Reference struct {
	t      ReferenceType
	n      ReferenceName
	h      Hash
	target ReferenceName
}

// ReferenceIter is a generic closable interface for iterating over References.
type ReferenceIter interface {
	Next() (*Reference, error)
	ForEach(func(*Reference) error) error
	Close()
}

// Hash return the hash of a hash reference
func (r *Reference) Hash() Hash {
	return r.h
}

// Name returns the Name of the Reference
func (r *Reference) Name() ReferenceName {
	return r.n
}

func (r ReferenceName) String() string {
	return string(r)
}

type sliceRefIter struct {
	pos    int
	series []Reference
}

func (iter *sliceRefIter) Next() (*Reference, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}

	obj := iter.series[iter.pos]
	iter.pos++
	return &obj, nil
}

func (iter *sliceRefIter) Close() {
	iter.pos = len(iter.series)
}

func (iter *sliceRefIter) ForEach(fn func(*Reference) error) error {
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
