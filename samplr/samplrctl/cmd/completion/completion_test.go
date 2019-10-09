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

package completion

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestGeneratesBash(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{
		Use: "root",
	}
	output := new(bytes.Buffer)
	root.SetOutput(output)
	root.SetArgs([]string{"completion", "bash"})

	AddCommand(ctx, root)

	err := root.Execute()
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
}

func TestErrorsOnNoShell(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{
		Use:           "root",
		SilenceErrors: false,
		SilenceUsage:  true,
	}
	output := new(bytes.Buffer)
	root.SetOutput(output)
	root.SetArgs([]string{"completion"})

	AddCommand(ctx, root)

	err := root.Execute()

	if err == nil {
		t.Error("Expected an error. Got: nil")
	} else if err != errUnspecifiedShell {
		t.Errorf("Expected Unspecified Shell error. Got %v", err)
	}
}

func TestErrorsOnUnsupportedShell(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{
		Use: "root",
	}
	output := new(bytes.Buffer)
	root.SetOutput(output)
	root.SetArgs([]string{"completion", "zsh"})

	AddCommand(ctx, root)

	err := root.Execute()
	if err == nil {
		t.Error("Expected an error. Got: nil")
	} else if err, ok := err.(*unsupportedShellErr); !ok {
		t.Errorf("Expected an unsupportedShellErr. Got %v", err)
	} else if err.shellName != "zsh" {
		t.Errorf("Expected err.shellName to be zsh. Got %v", err.shellName)
	}
}

func TestErrorsOnTooManyArgs(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{
		Use: "root",
	}
	output := new(bytes.Buffer)
	root.SetOutput(output)
	root.SetArgs([]string{"completion", "bash", "zsh", "fish"})

	AddCommand(ctx, root)

	err := root.Execute()
	if err == nil {
		t.Error("Expected an error. Got: nil")
	} else if err != errTooManyArgs {
		t.Errorf("Expected errTooManyArgs. Got %v", err)
	}
}
