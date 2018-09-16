package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCloudConfig struct {
	Hostname string `yaml:"hostname,omitempty"`
	Key1     string `yaml:"key1,omitempty"`
	Key2     string `yaml:"key2,omitempty"`
}

func TestConvertMergesLeftIntoRight(t *testing.T) {
	assert := require.New(t)
	cc0 := testCloudConfig{Key1: "k1v0", Key2: "k2v0"}
	cc1 := map[interface{}]interface{}{"key1": "k1value1", "hostname": "somehost"}
	Convert(cc1, &cc0)
	expected := testCloudConfig{Hostname: "somehost", Key1: "k1value1", Key2: "k2v0"}
	assert.Equal(expected, cc0)
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

func TestMerge(t *testing.T) {
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
	assert.Equal(expected, Merge(m0, m1))
}

func TestCmdLineStr(t *testing.T) {
	assert := require.New(t)

	cmdLine := `rancher.cloud_init.datasources=[\'url:http://192.168.1.100/cloud-config\']`
	assert.Equal("rancher.cloud_init.datasources=['url:http://192.168.1.100/cloud-config']", UnescapeKernelParams(cmdLine))

	cmdLine = `rancher.cloud_init.datasources=[\"url:http://192.168.1.100/cloud-config\"]`
	assert.Equal(`rancher.cloud_init.datasources=["url:http://192.168.1.100/cloud-config"]`, UnescapeKernelParams(cmdLine))
}
