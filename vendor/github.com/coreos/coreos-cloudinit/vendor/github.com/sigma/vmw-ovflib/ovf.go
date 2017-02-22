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
