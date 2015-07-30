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

// Etcd2 is a top-level structure which embeds its underlying configuration,
// config.Etcd2, and provides the system-specific Unit().
type Etcd2 struct {
	config.Etcd2
}

// Units creates a Unit file drop-in for etcd, using any configured options.
func (ee Etcd2) Units() []Unit {
	return []Unit{{config.Unit{
		Name:    "etcd2.service",
		Runtime: true,
		DropIns: []config.UnitDropIn{{
			Name:    "20-cloudinit.conf",
			Content: serviceContents(ee.Etcd2),
		}},
	}}}
}
