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

package config

type User struct {
	Name                 string   `yaml:"name"`
	PasswordHash         string   `yaml:"passwd"`
	SSHAuthorizedKeys    []string `yaml:"ssh_authorized_keys"`
	SSHImportGithubUser  string   `yaml:"coreos_ssh_import_github"       deprecated:"trying to fetch from a remote endpoint introduces too many intermittent errors"`
	SSHImportGithubUsers []string `yaml:"coreos_ssh_import_github_users" deprecated:"trying to fetch from a remote endpoint introduces too many intermittent errors"`
	SSHImportURL         string   `yaml:"coreos_ssh_import_url"          deprecated:"trying to fetch from a remote endpoint introduces too many intermittent errors"`
	GECOS                string   `yaml:"gecos"`
	Homedir              string   `yaml:"homedir"`
	NoCreateHome         bool     `yaml:"no_create_home"`
	PrimaryGroup         string   `yaml:"primary_group"`
	Groups               []string `yaml:"groups"`
	NoUserGroup          bool     `yaml:"no_user_group"`
	System               bool     `yaml:"system"`
	NoLogInit            bool     `yaml:"no_log_init"`
	Shell                string   `yaml:"shell"`
}
