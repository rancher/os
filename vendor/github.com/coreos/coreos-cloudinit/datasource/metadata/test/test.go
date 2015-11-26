// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type HttpClient struct {
	Resources map[string]string
	Err       error
}

func (t *HttpClient) GetRetry(url string) ([]byte, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	if val, ok := t.Resources[url]; ok {
		return []byte(val), nil
	} else {
		return nil, pkg.ErrNotFound{fmt.Errorf("not found: %q", url)}
	}
}

func (t *HttpClient) Get(url string) ([]byte, error) {
	return t.GetRetry(url)
}
