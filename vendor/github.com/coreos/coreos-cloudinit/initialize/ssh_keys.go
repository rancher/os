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

package initialize

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
)

type UserKey struct {
	ID  int    `json:"id,omitempty"`
	Key string `json:"key"`
}

func SSHImportKeysFromURL(system_user string, url string) error {
	keys, err := fetchUserKeys(url)
	if err != nil {
		return err
	}

	key_name := fmt.Sprintf("coreos-cloudinit-%s", system_user)
	return system.AuthorizeSSHKeys(system_user, key_name, keys)
}

func fetchUserKeys(url string) ([]string, error) {
	client := pkg.NewHttpClient()
	data, err := client.GetRetry(url)
	if err != nil {
		return nil, err
	}

	var userKeys []UserKey
	err = json.Unmarshal(data, &userKeys)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	for _, key := range userKeys {
		keys = append(keys, key.Key)
	}
	return keys, err
}
