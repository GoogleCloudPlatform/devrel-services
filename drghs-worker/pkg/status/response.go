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

package status

import (
	"encoding/json"
	"fmt"
	"io"
)

// Response is for API responses
type Response struct {
	Error  string    `json:",omitempty"`
	Issues []*Status `json:",omitempty"`
	Issue  *Status   `json:",omitempty"`
}

// WriteTo writes the response to an io.Writer
func (r Response) WriteTo(w io.Writer) (int64, error) {
	out, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		nerr := fmt.Errorf(`{"Error": %q}`, "could not marshal json: "+err.Error())
		io.WriteString(w, nerr.Error())
		return 0, err
	}
	i, err := w.Write(out)
	return int64(i), err
}
