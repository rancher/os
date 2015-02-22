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

package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// Add the provide SSH public key to the core user's list of
// authorized keys
func AuthorizeSSHKeys(user string, keysName string, keys []string) error {
	for _, key := range keys {
		trimmedKey := strings.TrimSpace(key)
		cmd := exec.Command("update-ssh-keys", user, trimmedKey)

		err := cmd.Start()
		if err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("Call to update-ssh-keys failed")
		}
	}

	return nil
}
