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

package network

import (
	"log"
	"strings"
)

func ProcessDebianNetconf(config []byte) ([]InterfaceGenerator, error) {
	log.Println("Processing Debian network config")
	lines := formatConfig(string(config))
	stanzas, err := parseStanzas(lines)
	if err != nil {
		return nil, err
	}

	interfaces := make([]*stanzaInterface, 0, len(stanzas))
	for _, stanza := range stanzas {
		switch s := stanza.(type) {
		case *stanzaInterface:
			interfaces = append(interfaces, s)
		}
	}
	log.Printf("Parsed %d network interfaces\n", len(interfaces))

	log.Println("Processed Debian network config")
	return buildInterfaces(interfaces), nil
}

func formatConfig(config string) []string {
	lines := []string{}
	config = strings.Replace(config, "\\\n", "", -1)
	for config != "" {
		split := strings.SplitN(config, "\n", 2)
		line := strings.TrimSpace(split[0])

		if len(split) == 2 {
			config = split[1]
		} else {
			config = ""
		}

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		lines = append(lines, line)
	}
	return lines
}
