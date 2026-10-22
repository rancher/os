package util

import (
	"os"
	"path/filepath"
	"strings"
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

const testContainerID = "8d24a1e2a5f26ca4b8f70a44a8f60ea2f2d3bdb2b3d3f9c1e0a5b4c3d2e1f009"

func TestParseCgroupV1(t *testing.T) {
	assert := require.New(t)

	cgroup := `12:blkio:/docker/` + testContainerID + `
11:pids:/docker/` + testContainerID + `
1:name=systemd:/docker/` + testContainerID + `
`
	id, err := parseCgroup(strings.NewReader(cgroup))
	assert.NoError(err)
	assert.Equal(testContainerID, id)
}

func TestParseCgroupV2(t *testing.T) {
	assert := require.New(t)

	id, err := parseCgroup(strings.NewReader("0::/docker/" + testContainerID + "\n"))
	assert.NoError(err)
	assert.Equal(testContainerID, id)

	id, err = parseCgroup(strings.NewReader("0::/system.slice/docker-" + testContainerID + ".scope\n"))
	assert.NoError(err)
	assert.Equal(testContainerID, id)
}

// System Docker does not create a cgroup per container on a cgroup v2 only
// system, so the container ID cannot be found in /proc/self/cgroup.
func TestParseCgroupWithoutContainerCgroup(t *testing.T) {
	assert := require.New(t)

	id, err := parseCgroup(strings.NewReader("0::/\n"))
	assert.NoError(err)
	assert.Equal("", id)
}

func TestParseMountInfo(t *testing.T) {
	assert := require.New(t)

	mountInfo := `525 480 0:135 / / rw,relatime master:180 - overlay overlay rw,lowerdir=/var/lib/system-docker/overlay2/l/YKAJ
530 525 0:139 / /sys/fs/cgroup rw,nosuid,nodev,noexec,relatime - cgroup2 cgroup2 rw,nsdelegate
545 525 8:1 /var/lib/system-docker/containers/` + testContainerID + `/resolv.conf /etc/resolv.conf rw,relatime - ext4 /dev/sda1 rw
546 525 8:1 /var/lib/system-docker/containers/` + testContainerID + `/hostname /etc/hostname rw,relatime - ext4 /dev/sda1 rw
547 525 8:1 /var/lib/docker /var/lib/docker rw,relatime - ext4 /dev/sda1 rw
`
	id, err := parseMountInfo(strings.NewReader(mountInfo))
	assert.NoError(err)
	assert.Equal(testContainerID, id)
}

func TestParseMountInfoWithoutEtcBindMounts(t *testing.T) {
	assert := require.New(t)

	mountInfo := `525 480 0:135 / / rw,relatime master:180 - overlay overlay rw,lowerdir=/var/lib/system-docker/overlay2/l/YKAJ
532 525 0:140 /var/lib/system-docker/containers/` + testContainerID + `/mounts/shm /dev/shm rw,nosuid,nodev,noexec,relatime - tmpfs shm rw,size=65536k
`
	id, err := parseMountInfo(strings.NewReader(mountInfo))
	assert.NoError(err)
	assert.Equal(testContainerID, id)
}

func TestParseMountInfoOutsideContainer(t *testing.T) {
	assert := require.New(t)

	mountInfo := `22 27 0:21 / /proc rw,relatime - proc none rw
27 0 8:1 / / rw,relatime - ext4 /dev/sda1 rw
`
	id, err := parseMountInfo(strings.NewReader(mountInfo))
	assert.NoError(err)
	assert.Equal("", id)
}

func TestFileCopyKeepsMode(t *testing.T) {
	assert := require.New(t)
	dir := t.TempDir()
	src, dest := filepath.Join(dir, "src"), filepath.Join(dir, "dest")
	assert.NoError(os.WriteFile(src, []byte("x"), 0644))
	assert.NoError(FileCopy(src, dest))
	info, err := os.Stat(dest)
	assert.NoError(err)
	assert.Equal(os.FileMode(0644), info.Mode().Perm())
}
