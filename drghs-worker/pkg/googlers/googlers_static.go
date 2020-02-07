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

package googlers

import "strings"

// GooglersStatic stores a list of Googlers in a hard-coded list
type GooglersStatic struct {
	// map from github login to ldap.
	googlers map[string]bool
}

// IsGoogler checks if the given username is a Googler or not.
func (s *GooglersStatic) IsGoogler(user string) bool {
	_, ok := s.googlers[user]
	return ok
}

// Update indicates the repository to update
func (s *GooglersStatic) Update() {
	return
}

// NewGooglersStatic instantiates and returns a new GooglersStatic struct
func NewGooglersStatic() *GooglersStatic {
	googlers := make(map[string]bool)
	lines := strings.Split(users, "\n")
	for _, line := range lines {
		googlers[line] = true
	}
	return &GooglersStatic{
		googlers: googlers,
	}
}

const users = `
surferjeffatgoogle
pfritzsche
michaelawyu
jadekler
tmatsuo
jba
andrewsg
jonparrott
tbpg
bshaffer
dzlier-gcp
ace-n
jabubake
kurtisvg
jsimonweb
ryanmats
frankyn
broady
lesv
afitz0
sanche21
simonz130
djmailhot
ahmetb
jmdobry
annie29
sgreenberg
davidcavazos
nkashy1
dizcology
alixhami
happyhuman
sirtorry
gguuss
remi
tswast
elibixby
nnegrey
puneith
oalami
jeffmendoza
rakyll
jerjou
p-buse
waprin
stephenplusplus
`
