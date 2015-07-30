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
	"log"
	"os/exec"
	"os/user"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

func UserExists(u *config.User) bool {
	_, err := user.Lookup(u.Name)
	return err == nil
}

func CreateUser(u *config.User) error {
	args := []string{}

	if u.PasswordHash != "" {
		args = append(args, "--password", u.PasswordHash)
	} else {
		args = append(args, "--password", "*")
	}

	if u.GECOS != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", u.GECOS))
	}

	if u.Homedir != "" {
		args = append(args, "--home-dir", u.Homedir)
	}

	if u.NoCreateHome {
		args = append(args, "--no-create-home")
	} else {
		args = append(args, "--create-home")
	}

	if u.PrimaryGroup != "" {
		args = append(args, "--gid", u.PrimaryGroup)
	}

	if len(u.Groups) > 0 {
		args = append(args, "--groups", strings.Join(u.Groups, ","))
	}

	if u.NoUserGroup {
		args = append(args, "--no-user-group")
	}

	if u.System {
		args = append(args, "--system")
	}

	if u.NoLogInit {
		args = append(args, "--no-log-init")
	}

	if u.Shell != "" {
		args = append(args, "--shell", u.Shell)
	}

	args = append(args, u.Name)

	output, err := exec.Command("useradd", args...).CombinedOutput()
	if err != nil {
		log.Printf("Command 'useradd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return err
}

func SetUserPassword(user, hash string) error {
	cmd := exec.Command("/usr/sbin/chpasswd", "-e")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	arg := fmt.Sprintf("%s:%s", user, hash)
	_, err = stdin.Write([]byte(arg))
	if err != nil {
		return err
	}
	stdin.Close()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
