// Copyright 2014-2015 VMware, Inc. All Rights Reserved.
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

package ovf

import (
	"encoding/xml"
	"log"
)

type environment struct {
	Platform   platform   `xml:"PlatformSection"`
	Properties []property `xml:"PropertySection>Property"`
}

type platform struct {
	Kind    string `xml:"Kind"`
	Version string `xml:"Version"`
	Vendor  string `xml:"Vendor"`
	Locale  string `xml:"Locale"`
}

type property struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type OvfEnvironment struct {
	Platform   platform
	Properties map[string]string
}

func ReadEnvironment(doc []byte) *OvfEnvironment {
	var env environment
	if err := xml.Unmarshal(doc, &env); err != nil {
		log.Fatalln(err)
	}

	dict := make(map[string]string)
	for _, p := range env.Properties {
		dict[p.Key] = p.Value
	}
	return &OvfEnvironment{Properties: dict,
		Platform: env.Platform}
}
