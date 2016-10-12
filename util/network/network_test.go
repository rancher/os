package network

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func NoTestLoadResourceSimple(t *testing.T) {
	assert := require.New(t)

	expected := `services:
- debian-console
- ubuntu-console
`
	expected = strings.TrimSpace(expected)

	b, e := LoadResource("https://raw.githubusercontent.com/rancher/os-services/v0.3.4/index.yml", true, false)

	assert.Nil(e)
	assert.Equal(expected, strings.TrimSpace(string(b)))
}
