package config

import (
	"fmt"
	"strings"
	"testing"
)

func testContains(t *testing.T, s string, substrs ...string) {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			t.Fail()
		}
	}
}

func TestGenerateEngineOptsString(t *testing.T) {
	if len(generateEngineOptsSlice(EngineOpts{})) != 0 {
		t.Fail()
	}
	if len(generateEngineOptsSlice(EngineOpts{
		Host: []string{
			"",
		},
	})) != 0 {
		t.Fail()
	}
	if len(generateEngineOptsSlice(EngineOpts{
		LogOpts: map[string]string{
			"max-file": "",
		},
	})) != 0 {
		t.Fail()
	}

	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		Bridge: "bridge",
	})), "--bridge bridge")

	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		SelinuxEnabled: &[]bool{true}[0],
	})), "--selinux-enabled")
	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		SelinuxEnabled: &[]bool{false}[0],
	})), "--selinux-enabled=false")

	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		Host: []string{
			"unix:///var/run/system-docker.sock",
			"unix:///var/run/docker.sock",
		},
	})), "--host unix:///var/run/system-docker.sock", "--host unix:///var/run/docker.sock")

	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		LogOpts: map[string]string{
			"max-size": "25m",
			"max-file": "2",
		},
	})), "--log-opt max-size=25m", "--log-opt max-file=2")

	testContains(t, fmt.Sprint(generateEngineOptsSlice(EngineOpts{
		Bridge:         "bridge",
		SelinuxEnabled: &[]bool{true}[0],
		LogOpts: map[string]string{
			"max-size": "25m",
			"max-file": "2",
		},
	})), "--bridge bridge", "--selinux-enabled", "--log-opt max-size=25m", "--log-opt max-file=2")
}
