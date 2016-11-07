package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	osYaml "github.com/rancher/os/config/yaml"

	log "github.com/Sirupsen/logrus"
)

const (
	DOCKER_CGROUPS_FILE = "/proc/self/cgroup"
)

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
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return WriteFileAtomic(dest, data, 0666)
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
		log.WithFields(log.Fields{"from": from, "err": err}).Warn("Error serializing to YML")
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

	if err := decoder.Decode(to); err != nil {
		return err
	}

	return nil
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

func RemoveString(slice []string, s string) []string {
	result := []string{}
	for _, elem := range slice {
		if elem != s {
			result = append(result, elem)
		}
	}
	return result
}

func ToStrings(data []interface{}) []string {
	result := make([]string, len(data), len(data))
	for k, v := range data {
		result[k] = v.(string)
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

func GetCurrentContainerId() (string, error) {
	file, err := os.Open(DOCKER_CGROUPS_FILE)

	if err != nil {
		return "", err
	}

	fileReader := bufio.NewScanner(file)
	if !fileReader.Scan() {
		return "", errors.New("Empty file /proc/self/cgroup")
	}
	line := fileReader.Text()
	parts := strings.Split(line, "/")

	for len(parts) != 3 {
		if !fileReader.Scan() {
			return "", errors.New("Found no docker cgroups")
		}
		line = fileReader.Text()
		parts = strings.Split(line, "/")
		if len(parts) == 3 {
			if strings.HasSuffix(parts[1], "docker") {
				break
			} else {
				parts = nil
			}
		}
	}

	return parts[len(parts)-1:][0], nil
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

func RunScript(path string) error {
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
		cmd = exec.Command(path)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunCommandSequence(commandSequence []osYaml.StringandSlice) {
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
			log.Errorf("Failed to run %s: %v", command, err)
		}
	}
}
