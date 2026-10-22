package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	osYaml "github.com/burmilla/os/config/yaml"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

const (
	cgroupFile    = "/proc/self/cgroup"
	mountInfoFile = "/proc/self/mountinfo"

	containerIDLength = 64
)

// Files which Docker bind mounts into every container from
// <docker-root>/containers/<container-id>/.
var containerMountPoints = []string{"/etc/resolv.conf", "/etc/hostname", "/etc/hosts"}

type AnyMap map[interface{}]interface{}

func Contains(values []string, value string) bool {
	if len(value) == 0 {
		return false
	}

	for _, i := range values {
		if i == value {
			return true
		}
	}

	return false
}

type ReturnsErr func() error

func FileCopy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return WriteFileAtomic(dest, data, info.Mode().Perm())
}

func HTTPDownloadToFile(url, dest string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// only caller downloads a cloud-config, which can hold secrets
	return WriteFileAtomic(dest, body, 0600)
}

func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir, file := path.Split(filename)
	tempFile, err := ioutil.TempFile(dir, fmt.Sprintf(".%s", file))
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(data); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempFile.Name(), perm); err != nil {
		return err
	}

	return os.Rename(tempFile.Name(), filename)
}

func Convert(from, to interface{}) error {
	bytes, err := yaml.Marshal(from)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, to)
}

func ConvertIgnoreOmitEmpty(from, to interface{}) error {
	var buffer bytes.Buffer

	encoder := yaml.NewEncoder(&buffer)
	encoder.IgnoreOmitEmpty = true

	if err := encoder.Encode(from); err != nil {
		return err
	}

	decoder := yaml.NewDecoder(&buffer)

	return decoder.Decode(to)
}

func Copy(d interface{}) interface{} {
	switch d := d.(type) {
	case map[interface{}]interface{}:
		return MapCopy(d)
	case []interface{}:
		return SliceCopy(d)
	default:
		return d
	}
}

func Merge(left, right map[interface{}]interface{}) map[interface{}]interface{} {
	result := MapCopy(left)

	for k, r := range right {
		if l, ok := left[k]; ok {
			switch l := l.(type) {
			case map[interface{}]interface{}:
				switch r := r.(type) {
				case map[interface{}]interface{}:
					result[k] = Merge(l, r)
				default:
					result[k] = r
				}
			default:
				result[k] = r
			}
		} else {
			result[k] = Copy(r)
		}
	}

	return result
}

func MapCopy(data map[interface{}]interface{}) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}
	for k, v := range data {
		result[k] = Copy(v)
	}
	return result
}

func SliceCopy(data []interface{}) []interface{} {
	result := make([]interface{}, len(data), len(data))
	for k, v := range data {
		result[k] = Copy(v)
	}
	return result
}

func Map2KVPairs(m map[string]string) []string {
	r := make([]string, 0, len(m))
	for k, v := range m {
		r = append(r, k+"="+v)
	}
	return r
}

func KVPairs2Map(kvs []string) map[string]string {
	r := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		s := strings.SplitN(kv, "=", 2)
		r[s[0]] = s[1]
	}
	return r
}

func TrimSplitN(str, sep string, count int) []string {
	result := []string{}
	for _, part := range strings.SplitN(strings.TrimSpace(str), sep, count) {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

func TrimSplit(str, sep string) []string {
	return TrimSplitN(str, sep, -1)
}

// GetCurrentContainerID returns the ID of the Docker container this process is
// running in.
func GetCurrentContainerID() (string, error) {
	id, err := containerIDFromCgroup()
	if err != nil {
		return "", err
	}
	if id != "" {
		return id, nil
	}

	// Docker only names the cgroup of a container after the container itself
	// when it manages that cgroup. System Docker does not do so on a cgroup v2
	// only system, which leaves us with a plain "0::/", so fall back to the
	// per container files which are bind mounted into every container.
	id, err = containerIDFromMountInfo()
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", errors.New("Found no docker container ID for the current process")
	}

	return id, nil
}

// containerIDFromCgroup looks for the container ID in the cgroup paths of the
// current process. It returns an empty string when none of them names a
// container, which is the normal case outside of a container and on a cgroup
// v2 only system.
func containerIDFromCgroup() (string, error) {
	file, err := os.Open(cgroupFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return parseCgroup(file)
}

func parseCgroup(cgroup io.Reader) (string, error) {
	scanner := bufio.NewScanner(cgroup)
	for scanner.Scan() {
		// hierarchy-ID:controller-list:cgroup-path, where the cgroup v2
		// unified hierarchy uses ID 0 and an empty controller list.
		fields := strings.SplitN(scanner.Text(), ":", 3)
		if len(fields) != 3 {
			continue
		}

		// The container ID is either a path element of its own, as in
		// /docker/<id>, or part of a scope, as in
		// /system.slice/docker-<id>.scope.
		for _, part := range strings.Split(fields[2], "/") {
			part = strings.TrimPrefix(strings.TrimSuffix(part, ".scope"), "docker-")
			if isContainerID(part) {
				return part, nil
			}
		}
	}

	return "", scanner.Err()
}

// containerIDFromMountInfo looks for the container ID in the source paths of
// the files which Docker bind mounts into every container. It returns an empty
// string when the current process is not running in a container.
func containerIDFromMountInfo() (string, error) {
	file, err := os.Open(mountInfoFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return parseMountInfo(file)
}

func parseMountInfo(mountInfo io.Reader) (string, error) {
	id := ""
	otherID := ""

	scanner := bufio.NewScanner(mountInfo)
	for scanner.Scan() {
		// The fourth field is the root of the mount within its filesystem and
		// the fifth one is the mount point, for example:
		// 273 259 8:1 /var/lib/system-docker/containers/<id>/resolv.conf /etc/resolv.conf rw,relatime - ext4 ...
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}

		found := containerIDFromPath(fields[3])
		if found == "" {
			continue
		}

		// Later mounts shadow earlier ones, so keep looking to end up with
		// the file which is actually visible on the mount point.
		if Contains(containerMountPoints, fields[4]) {
			id = found
		} else {
			// Anything else mounted out of the container directory, such as
			// /dev/shm, still identifies the container.
			otherID = found
		}
	}

	if id == "" {
		id = otherID
	}

	return id, scanner.Err()
}

// containerIDFromPath returns the container ID of a <docker-root>/containers/
// <container-id>/<file> path, or an empty string when path is not one.
func containerIDFromPath(mountRoot string) string {
	parts := strings.Split(mountRoot, "/")
	for i, part := range parts {
		if part == "containers" && i+1 < len(parts) && isContainerID(parts[i+1]) {
			return parts[i+1]
		}
	}

	return ""
}

func isContainerID(s string) bool {
	if len(s) != containerIDLength {
		return false
	}

	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}

	return true
}

func UnescapeKernelParams(s string) string {
	s = strings.Replace(s, `\"`, `"`, -1)
	s = strings.Replace(s, `\'`, `'`, -1)
	return s
}

func ExistsAndExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := info.Mode().Perm()
	return mode&os.ModePerm != 0
}

func RunScript(path string, args ...string) error {
	if !ExistsAndExecutable(path) {
		return nil
	}

	script, err := os.Open(path)
	if err != nil {
		return err
	}

	magic := make([]byte, 2)
	if _, err = script.Read(magic); err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", path)
	if string(magic) == "#!" {
		cmd = exec.Command(path, args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunCommandSequence(commandSequence []osYaml.StringandSlice) error {
	for _, command := range commandSequence {
		var cmd *exec.Cmd
		if command.StringValue != "" {
			cmd = exec.Command("sh", "-c", command.StringValue)
		} else if len(command.SliceValue) > 0 {
			cmd = exec.Command(command.SliceValue[0], command.SliceValue[1:]...)
		} else {
			continue
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run %s: %v", command, err)
		}
	}
	return nil
}

func GenerateDindEngineScript(name string) error {
	if err := RemoveDindEngineScript(name); err != nil {
		return err
	}

	bytes := []byte("/usr/bin/docker -H unix:///var/lib/m-user-docker/" + name + "/docker-" + name + ".sock $@")

	err := ioutil.WriteFile("/usr/bin/docker-"+name, bytes, 0755)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDindEngineScript(name string) error {
	if _, err := os.Stat("/usr/bin/docker-" + name); err == nil {
		err = os.Remove("/usr/bin/docker-" + name)
		if err != nil {
			return err
		}
	}
	return nil
}
