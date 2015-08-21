package util

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNilMap(t *testing.T) {
	assert := require.New(t)
	var m map[string]interface{} = nil
	assert.True(m == nil)
}

func TestMapCopy(t *testing.T) {
	assert := require.New(t)
	m0 := map[interface{}]interface{}{"a": 1, "b": map[interface{}]interface{}{"c": 3}, "d": "4"}
	m1 := MapCopy(m0)
	assert.Equal(m0, m1)

	delete(m0, "a")
	assert.Equal(len(m1), len(m0)+1)

	b0 := m0["b"].(map[interface{}]interface{})
	b1 := m1["b"].(map[interface{}]interface{})
	b1["e"] = "queer"

	assert.Equal(len(b1), len(b0)+1)
}

func TestSliceCopy(t *testing.T) {
	assert := require.New(t)
	m0 := []interface{}{1, map[interface{}]interface{}{"c": 3}, "4"}
	m1 := SliceCopy(m0)
	assert.Equal(m0, m1)

	m0 = m0[1:]
	assert.Equal(len(m1), len(m0)+1)

	b0 := m0[0].(map[interface{}]interface{})
	b1 := m1[1].(map[interface{}]interface{})
	b1["e"] = "queer"

	assert.Equal(len(b1), len(b0)+1)
}

func TestMapsIntersection(t *testing.T) {
	assert := require.New(t)

	m0 := map[interface{}]interface{}{
		"a": 1,
		"b": map[interface{}]interface{}{"c": 3},
		"d": "4",
		"e": []interface{}{1, 2, 3},
	}
	m1 := MapCopy(m0)

	delete(m0, "a")
	b1 := m1["b"].(map[interface{}]interface{})
	delete(b1, "c")
	m1["e"] = []interface{}{2, 3, 4}
	expected := map[interface{}]interface{}{"b": map[interface{}]interface{}{}, "d": "4", "e": []interface{}{2, 3}}
	assert.Equal(expected, MapsIntersection(m0, m1, Equal))
}

func TestMapsUnion(t *testing.T) {
	assert := require.New(t)

	m0 := map[interface{}]interface{}{
		"a": 1,
		"b": map[interface{}]interface{}{"c": 3},
		"d": "4",
		"f": []interface{}{1, 2, 3},
	}
	m1 := MapCopy(m0)
	m1["e"] = "added"
	m1["d"] = "replaced"
	m1["f"] = []interface{}{2, 3, 4}

	delete(m0, "a")
	b1 := m1["b"].(map[interface{}]interface{})
	delete(b1, "c")
	expected := map[interface{}]interface{}{
		"a": 1,
		"b": map[interface{}]interface{}{"c": 3},
		"d": "replaced",
		"e": "added",
		"f": []interface{}{1, 2, 3, 4},
	}
	assert.Equal(expected, MapsUnion(m0, m1, Replace))
}

func TestLoadResourceSimple(t *testing.T) {
	assert := require.New(t)

	expected := `services:
- debian-console
- ubuntu-console
`
	expected = strings.TrimSpace(expected)

	b, e := LoadResource("https://raw.githubusercontent.com/rancherio/os-services/v0.3.4/index.yml", true, []string{})

	assert.Nil(e)
	assert.Equal(expected, strings.TrimSpace(string(b)))
}
