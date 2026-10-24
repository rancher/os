package compose

import "testing"

func TestResolveEngine(t *testing.T) {
	available := []string{"docker-25.0.14", "docker-29.10.0", "docker-29.6.1", "docker-29.9.2"}

	tests := []struct {
		name          string
		engine        string
		defaultEngine string
		available     []string
		want          string
		wantOK        bool
	}{
		{"available", "docker-29.6.1", "latest", available, "docker-29.6.1", true},
		{"missing falls back to latest", "docker-29.6.2", "latest", available, "docker-29.10.0", false},
		{"missing falls back to listed default", "docker-29.6.2", "docker-25.0.14", available, "docker-25.0.14", false},
		{"index not readable", "docker-29.6.2", "latest", nil, "docker-29.6.2", true},
		{"latest resolves to latest available", "latest", "latest", available, "docker-29.10.0", false},
		{"latest without index", "latest", "latest", nil, "latest", true},
		{"local file", "/var/lib/rancher/conf/docker.yml", "latest", available, "/var/lib/rancher/conf/docker.yml", true},
		{"url", "https://example.com/docker.yml", "latest", available, "https://example.com/docker.yml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveEngine(tt.engine, tt.defaultEngine, tt.available)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("resolveEngine(%q, %q, %v) = %q, %v; want %q, %v", tt.engine, tt.defaultEngine, tt.available, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
