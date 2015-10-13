package util

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type testCloudConfig struct {
	Hostname string `yaml:"hostname,omitempty"`
	Key1     string `yaml:"key1,omitempty"`
	Key2     string `yaml:"key2,omitempty"`
}

func TestPassByValue(t *testing.T) {
	assert := require.New(t)
	cc0ptr := &testCloudConfig{}
	cc0ptr.Hostname = "test0"
	cc1 := *cc0ptr
	cc1.Hostname = "test1"
	assert.NotEqual(cc0ptr.Hostname, cc1.Hostname)
}

func TestConvertMergesLeftIntoRight(t *testing.T) {
	assert := require.New(t)
	cc0 := testCloudConfig{Key1: "k1v0", Key2: "k2v0"}
	cc1 := map[interface{}]interface{}{"key1": "k1value1", "hostname": "somehost"}
	Convert(cc1, &cc0)
	expected := testCloudConfig{Hostname: "somehost", Key1: "k1value1", Key2: "k2v0"}
	assert.Equal(expected, cc0)
}

func TestNilMap(t *testing.T) {
	assert := require.New(t)
	var m map[string]interface{} = nil
	assert.True(m == nil)
	assert.True(len(m) == 0)
}

func NoTestCopyPointer(t *testing.T) {
	assert := require.New(t)
	testCCpt := &testCloudConfig{}
	m0 := map[string]interface{}{"a": testCCpt, "b": testCCpt}
	m1 := Copy(m0).(map[string]interface{})
	m1["a"].(*testCloudConfig).Hostname = "somehost"
	assert.Equal("", m0["a"].(*testCloudConfig).Hostname)
	assert.Equal("somehost", m1["a"].(*testCloudConfig).Hostname)
	assert.Equal("", m1["b"].(*testCloudConfig).Hostname)
}

func TestEmptyMap(t *testing.T) {
	assert := require.New(t)
	m := map[interface{}]interface{}{}
	assert.True(len(m) == 0)
}

func tryMutateArg(p *string) *string {
	s := "test"
	p = &s
	return p
}

func TestMutableArg(t *testing.T) {
	assert := require.New(t)
	s := "somestring"
	p := &s
	assert.NotEqual(tryMutateArg(p), p)
}

func TestFilter(t *testing.T) {
	assert := require.New(t)
	ss := []interface{}{"1", "2", "3", "4"}
	assert.Equal([]interface{}{"1", "2", "4"}, Filter(ss, func(x interface{}) bool { return x != "3" }))

	ss1 := append([]interface{}{}, "qqq")
	assert.Equal([]interface{}{"qqq"}, ss1)

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
	expected := map[interface{}]interface{}{"b": map[interface{}]interface{}{}, "d": "4"}
	assert.Equal(expected, MapsIntersection(m0, m1))
}

func TestMapsDifference(t *testing.T) {
	assert := require.New(t)

	m0 := map[interface{}]interface{}{
		"a": 1,
		"b": map[interface{}]interface{}{"c": 3},
		"d": "4",
		"e": []interface{}{1, 2, 3},
	}
	m1 := MapCopy(m0)

	assert.Equal(map[interface{}]interface{}{}, MapsDifference(m0, m0))
	assert.Equal(map[interface{}]interface{}{}, MapsDifference(m0, m1))

	delete(m1, "a")
	b1 := m1["b"].(map[interface{}]interface{})
	delete(b1, "c")
	m1["e"] = []interface{}{2, 3, 4}

	expectedM1M0 := map[interface{}]interface{}{"b": map[interface{}]interface{}{}, "e": []interface{}{2, 3, 4}}
	assert.Equal(expectedM1M0, MapsDifference(m1, m0))

	expectedM0M1 := map[interface{}]interface{}{"a": 1, "b": map[interface{}]interface{}{"c": 3}, "e": []interface{}{1, 2, 3}}
	assert.Equal(expectedM0M1, MapsDifference(m0, m1))
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
		"f": []interface{}{2, 3, 4},
	}
	assert.Equal(expected, MapsUnion(m0, m1))
}

func NoTestLoadResourceSimple(t *testing.T) {
	assert := require.New(t)

	expected := `services:
- debian-console
- ubuntu-console
`
	expected = strings.TrimSpace(expected)

	b, e := LoadResource("https://raw.githubusercontent.com/rancher/os-services/v0.3.4/index.yml", true, []string{})

	assert.Nil(e)
	assert.Equal(expected, strings.TrimSpace(string(b)))
}
