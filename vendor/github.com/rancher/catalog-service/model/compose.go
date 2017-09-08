package model

type RancherCompose struct {
	Name                  string            `yaml:"name"`
	UUID                  string            `yaml:"uuid"`
	Description           string            `yaml:"description"`
	Version               string            `yaml:"version"`
	Questions             []Question        `json:"questions" yaml:"questions,omitempty"`
	MinimumRancherVersion string            `json:"minimumRancherVersion" yaml:"minimum_rancher_version,omitempty"`
	MaximumRancherVersion string            `json:"maximumRancherVersion" yaml:"maximum_rancher_version,omitempty"`
	Labels                map[string]string `json:"labels" yaml:"labels,omitempty"`
	UpgradeFrom           string            `json:"upgradeFrom" yaml:"upgrade_from,omitempty"`
}
