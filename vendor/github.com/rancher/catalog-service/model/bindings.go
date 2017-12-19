package model

type Bindings struct {
	Labels map[string]string `json:"labels"`
	Ports  []string          `json:"ports"`
}
