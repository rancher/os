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
	"net"
	"os"
	"testing"

	"github.com/burmilla/os/config/cloudinit/datasource"
)

func TestEnvironmentApply(t *testing.T) {
	os.Setenv("COREOS_PUBLIC_IPV4", "1.2.3.4")
	os.Setenv("COREOS_PRIVATE_IPV4", "5.6.7.8")
	os.Setenv("COREOS_PUBLIC_IPV6", "1234::")
	os.Setenv("COREOS_PRIVATE_IPV6", "5678::")
	for _, tt := range []struct {
		metadata datasource.Metadata
		input    string
		out      string
	}{
		{
			// Substituting both values directly should always take precedence
			// over environment variables
			datasource.Metadata{
				PublicIPv4:  net.ParseIP("192.0.2.3"),
				PrivateIPv4: net.ParseIP("192.0.2.203"),
				PublicIPv6:  net.ParseIP("fe00:1234::"),
				PrivateIPv6: net.ParseIP("fe00:5678::"),
			},
			`[Service]
ExecStart=/usr/bin/echo "$public_ipv4 $public_ipv6"
ExecStop=/usr/bin/echo $private_ipv4 $private_ipv6
ExecStop=/usr/bin/echo $unknown`,
			`[Service]
ExecStart=/usr/bin/echo "192.0.2.3 fe00:1234::"
ExecStop=/usr/bin/echo 192.0.2.203 fe00:5678::
ExecStop=/usr/bin/echo $unknown`,
		},
		{
			// Substituting one value directly while falling back with the other
			datasource.Metadata{
				PrivateIPv4: net.ParseIP("127.0.0.1"),
			},
			"$private_ipv4\n$public_ipv4",
			"127.0.0.1\n1.2.3.4",
		},
		{
			// Falling back to environment variables for both values
			datasource.Metadata{},
			"$private_ipv4\n$public_ipv4",
			"5.6.7.8\n1.2.3.4",
		},
		{
			// No substitutions
			datasource.Metadata{},
			"$private_ipv4\nfoobar",
			"5.6.7.8\nfoobar",
		},
		{
			// Escaping substitutions
			datasource.Metadata{
				PrivateIPv4: net.ParseIP("127.0.0.1"),
			},
			`\$private_ipv4
$private_ipv4
addr: \$private_ipv4
\\$private_ipv4`,
			`$private_ipv4
127.0.0.1
addr: $private_ipv4
\$private_ipv4`,
		},
		{
			// No substitutions with escaping
			datasource.Metadata{},
			"\\$test\n$test",
			"\\$test\n$test",
		},
	} {

		env := NewEnvironment(tt.metadata)
		got := env.Apply(tt.input)
		if got != tt.out {
			t.Fatalf("Environment incorrectly applied.\ngot:\n%s\nwant:\n%s", got, tt.out)
		}
	}
}
