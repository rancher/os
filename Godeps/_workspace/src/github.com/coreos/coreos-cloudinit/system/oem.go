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
	"path"

	"github.com/coreos/coreos-cloudinit/config"
)

// OEM is a top-level structure which embeds its underlying configuration,
// config.OEM, and provides the system-specific File().
type OEM struct {
	config.OEM
}

func (oem OEM) File() (*File, error) {
	if oem.ID == "" {
		return nil, nil
	}

	content := fmt.Sprintf("ID=%s\n", oem.ID)
	content += fmt.Sprintf("VERSION_ID=%s\n", oem.VersionID)
	content += fmt.Sprintf("NAME=%q\n", oem.Name)
	content += fmt.Sprintf("HOME_URL=%q\n", oem.HomeURL)
	content += fmt.Sprintf("BUG_REPORT_URL=%q\n", oem.BugReportURL)

	return &File{config.File{
		Path:               path.Join("etc", "oem-release"),
		RawFilePermissions: "0644",
		Content:            content,
	}}, nil
}
