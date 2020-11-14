// Copyright 2015 CoreOS, Inc.
// Copyright 2015-2017 Rancher Labs, Inc.
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

import (
	"net"
	"testing"

	"github.com/burmilla/os/config/cloudinit/datasource"
)

func TestSubstituteUserDataVars(t *testing.T) {
	for _, tt := range []struct {
		metadata datasource.Metadata
		input    string
		out      string
	}{
		{
			// Userdata with docker-compose syntax
			datasource.Metadata{
				PublicIPv4:  net.ParseIP("192.0.2.3"),
				PrivateIPv4: net.ParseIP("192.0.2.203"),
				PublicIPv6:  net.ParseIP("fe00:1234::"),
				PrivateIPv6: net.ParseIP("fe00:5678::"),
			},
			`servicexyz:
			  image: rancher/servicexyz:v0.3.1
			  ports:
			  - "$public_ipv4:8001:8001"
			  - "$public_ipv6:8001:8001"
			  - "$private_ipv4:8001:8001"
			  - "$private_ipv6:8001:8001"`,
			`servicexyz:
			  image: rancher/servicexyz:v0.3.1
			  ports:
			  - "192.0.2.3:8001:8001"
			  - "fe00:1234:::8001:8001"
			  - "192.0.2.203:8001:8001"
			  - "fe00:5678:::8001:8001"`,
		},
		{
			// Userdata with cloud-config/rancher syntax
			datasource.Metadata{
				PublicIPv4:  net.ParseIP("192.0.2.3"),
				PrivateIPv4: net.ParseIP("192.0.2.203"),
				PublicIPv6:  net.ParseIP("fe00:1234::"),
				PrivateIPv6: net.ParseIP("fe00:5678::"),
			},
			`write_files:
			    - path: /etc/environment
			      content: |
			        PRIVATE_IPV6=$private_ipv6
			        PUBLIC_IPV6=$public_ipv6
			rancher:
			  network:
			    interfaces:
			      eth1:
			        address: $private_ipv4/16
			  docker:
			  	tls_args: ['-H=$public_ipv4:2376']`,
			`write_files:
			    - path: /etc/environment
			      content: |
			        PRIVATE_IPV6=fe00:5678::
			        PUBLIC_IPV6=fe00:1234::
			rancher:
			  network:
			    interfaces:
			      eth1:
			        address: 192.0.2.203/16
			  docker:
			  	tls_args: ['-H=192.0.2.3:2376']`,
		},
		{
			// no metadata
			datasource.Metadata{},
			"address: $private_ipv4",
			"address: ",
		},
	} {

		got := substituteVars([]byte(tt.input), tt.metadata)
		if string(got) != tt.out {
			t.Fatalf("Userdata substitution incorrectly applied.\ngot:\n%s\nwant:\n%s", got, tt.out)
		}
	}
}
