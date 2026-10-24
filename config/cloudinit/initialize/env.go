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
	"regexp"
	"strings"

	"github.com/burmilla/os/config/cloudinit/datasource"
)

type Environment struct {
	substitutions map[string]string
}

func NewEnvironment(metadata datasource.Metadata) *Environment {
	firstNonNull := func(ip net.IP, env string) string {
		if ip == nil {
			return env
		}
		return ip.String()
	}
	substitutions := map[string]string{
		"$public_ipv4":  firstNonNull(metadata.PublicIPv4, os.Getenv("COREOS_PUBLIC_IPV4")),
		"$private_ipv4": firstNonNull(metadata.PrivateIPv4, os.Getenv("COREOS_PRIVATE_IPV4")),
		"$public_ipv6":  firstNonNull(metadata.PublicIPv6, os.Getenv("COREOS_PUBLIC_IPV6")),
		"$private_ipv6": firstNonNull(metadata.PrivateIPv6, os.Getenv("COREOS_PRIVATE_IPV6")),
	}
	return &Environment{substitutions}
}

// Apply goes through the map of substitutions and replaces all instances of
// the keys with their respective values. It supports escaping substitutions
// with a leading '\'.
func (e *Environment) Apply(data string) string {
	for key, val := range e.substitutions {
		matchKey := strings.Replace(key, `$`, `\$`, -1)
		replKey := strings.Replace(key, `$`, `$$`, -1)

		// "key" -> "val"
		data = regexp.MustCompile(`([^\\]|^)`+matchKey).ReplaceAllString(data, `${1}`+val)
		// "\key" -> "key"
		data = regexp.MustCompile(`\\`+matchKey).ReplaceAllString(data, replKey)
	}
	return data
}
