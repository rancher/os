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

package datasource

import (
	"net"

	"github.com/rancher/os/pkg/netconf"
)

type Datasource interface {
	IsAvailable() bool
	AvailabilityChanges() bool
	ConfigRoot() string
	FetchMetadata() (Metadata, error)
	FetchUserdata() ([]byte, error)
	Type() string
	String() string
	// Finish gives the datasource the opportunity to clean up, unmount or release any open / cache resources
	Finish() error
}

type Metadata struct {
	// TODO: move to netconf/types.go ?
	// see https://ahmetalpbalkan.com/blog/comparison-of-instance-metadata-services/
	Hostname      string
	SSHPublicKeys map[string]string
	NetworkConfig netconf.NetworkConfig
	RootDisk      string

	PublicIPv4  net.IP
	PublicIPv6  net.IP
	PrivateIPv4 net.IP
	PrivateIPv6 net.IP
}
