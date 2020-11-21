package control

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenTpl(t *testing.T) {
	assert := require.New(t)
	tpl := `
  services:
    {{if eq "amd64" .ARCH -}}
    acpid:
      image: burmilla/os-acpid:0.x.x
      labels:
        io.rancher.os.scope: system
      net: host
      uts: host
      privileged: true
      volumes_from:
      - command-volumes
      - system-volumes
    {{end -}}
    all-volumes:`

	for _, tc := range []struct {
		arch     string
		expected string
	}{
		{"amd64", `
  services:
    acpid:
      image: burmilla/os-acpid:0.x.x
      labels:
        io.rancher.os.scope: system
      net: host
      uts: host
      privileged: true
      volumes_from:
      - command-volumes
      - system-volumes
    all-volumes:`},
		{"arm", `
  services:
    all-volumes:`},
	} {
		out := &bytes.Buffer{}
		os.Setenv("ARCH", tc.arch)
		genTpl(strings.NewReader(tpl), out)
		assert.Equal(tc.expected, out.String(), tc.arch)
	}
}
