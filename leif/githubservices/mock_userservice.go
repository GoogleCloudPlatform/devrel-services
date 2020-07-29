// Copyright 2020 Google LLC
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

// This package was modeled after the mocking strategy outlined at:
// https://github.com/google/go-github/issues/113#issuecomment-46023864

package githubservices

import (
	"context"
	"errors"

	"github.com/google/go-github/github"
)

// MockGithubUserService is a struct that can replace github.UsersService for testing
type MockGithubUserService struct {
	Response *github.Response
	Error    error
	User     string
}

// Get mocks the original github.UsersService.Get()
// Checks whether the user is correct and returns the mocked response and error
func (mgc *MockGithubUserService) Get(ctx context.Context, user string) (*github.User, *github.Response, error) {
	if user != mgc.User {
		return nil, nil, errors.New("user did not equal expected user: was: " + user)
	}
	return nil, mgc.Response, mgc.Error
}
