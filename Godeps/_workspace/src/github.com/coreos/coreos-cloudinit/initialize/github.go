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
	"fmt"

	"github.com/coreos/coreos-cloudinit/system"
)

func SSHImportGithubUser(system_user string, github_user string) error {
	url := fmt.Sprintf("https://api.github.com/users/%s/keys", github_user)
	keys, err := fetchUserKeys(url)
	if err != nil {
		return err
	}

	key_name := fmt.Sprintf("github-%s", github_user)
	return system.AuthorizeSSHKeys(system_user, key_name, keys)
}
