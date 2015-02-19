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
	"github.com/coreos/coreos-cloudinit/config"
)

// Fleet is a top-level structure which embeds its underlying configuration,
// config.Fleet, and provides the system-specific Unit().
type Fleet struct {
	config.Fleet
}

// Units generates a Unit file drop-in for fleet, if any fleet options were
// configured in cloud-config
func (fe Fleet) Units() []Unit {
	return []Unit{{config.Unit{
		Name:    "fleet.service",
		Runtime: true,
		DropIns: []config.UnitDropIn{{
			Name:    "20-cloudinit.conf",
			Content: serviceContents(fe.Fleet),
		}},
	}}}
}
