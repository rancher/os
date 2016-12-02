package config

import (
	"fmt"
	"os"

	"github.com/fatih/structs"
)

func (d *DockerConfig) FullArgs() []string {
	args := []string{"daemon"}
	args = append(args, generateEngineOptsSlice(d.EngineOpts)...)
	args = append(args, d.ExtraArgs...)
	if d.TLS {
		args = append(args, d.TLSArgs...)
	}
	return args
}

func (d *DockerConfig) AppendEnv() []string {
	return append(os.Environ(), d.Environment...)
}

func generateEngineOptsSlice(opts EngineOpts) []string {
	optsStruct := structs.New(opts)

	var optsSlice []string
	for k, v := range optsStruct.Map() {
		optTag := optsStruct.Field(k).Tag("opt")

		switch value := v.(type) {
		case string:
			if value != "" {
				optsSlice = append(optsSlice, fmt.Sprintf("--%s", optTag), value)
			}
		case *bool:
			if value != nil {
				if *value {
					optsSlice = append(optsSlice, fmt.Sprintf("--%s", optTag))
				} else {
					optsSlice = append(optsSlice, fmt.Sprintf("--%s=false", optTag))
				}
			}
		case []string:
			for _, elem := range value {
				if elem != "" {
					optsSlice = append(optsSlice, fmt.Sprintf("--%s", optTag), elem)
				}
			}
		case map[string]string:
			for k, v := range value {
				if v != "" {
					optsSlice = append(optsSlice, fmt.Sprintf("--%s", optTag), fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
	}

	return optsSlice
}
