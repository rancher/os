package util

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/pkg/mount"
	"gopkg.in/yaml.v2"
)

var (
	letters      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	ErrNoNetwork = errors.New("Networking not available to load resource")
	ErrNotFound  = errors.New("Failed to find resource")
)

func mountProc() error {
	if _, err := os.Stat("/proc/self/mountinfo"); os.IsNotExist(err) {
		if _, err := os.Stat("/proc"); os.IsNotExist(err) {
			if err = os.Mkdir("/proc", 0755); err != nil {
				return err
			}
		}

		if err := syscall.Mount("none", "/proc", "proc", 0, ""); err != nil {
			return err
		}
	}

	return nil
}

func Mount(device, directory, fsType, options string) error {
	if err := mountProc(); err != nil {
		return nil
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	return mount.Mount(device, directory, fsType, options)
}

func Remount(directory, options string) error {
	return mount.Mount("", directory, "", fmt.Sprintf("remount,%s", options))
}

func ExtractTar(archive string, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	input := tar.NewReader(f)

	for {
		header, err := input.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if header == nil {
			break
		}

		fileInfo := header.FileInfo()
		fileName := path.Join(dest, header.Name)
		if fileInfo.IsDir() {
			//log.Debugf("DIR : %s", fileName)
			err = os.MkdirAll(fileName, fileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			//log.Debugf("FILE: %s", fileName)
			destFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(destFile, input)
			// Not deferring, concerned about holding open too many files
			destFile.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

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

func ShortCircuit(funcs ...ReturnsErr) error {
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}

type ErrWriter struct {
	w   io.Writer
	Err error
}

func NewErrorWriter(w io.Writer) *ErrWriter {
	return &ErrWriter{
		w: w,
	}
}

func (e *ErrWriter) Write(buf []byte) *ErrWriter {
	if e.Err != nil {
		return e
	}

	_, e.Err = e.w.Write(buf)
	return e
}

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Convert(from, to interface{}) error {
	bytes, err := yaml.Marshal(from)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, to)
}

func MergeBytes(left, right []byte) ([]byte, error) {
	leftMap := make(map[interface{}]interface{})
	rightMap := make(map[interface{}]interface{})

	err := yaml.Unmarshal(left, &leftMap)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(right, &rightMap)
	if err != nil {
		return nil, err
	}

	MergeMaps(leftMap, rightMap)

	return yaml.Marshal(leftMap)
}

func MergeMaps(left, right map[interface{}]interface{}) {
	for k, v := range right {
		merged := false
		if existing, ok := left[k]; ok {
			if rightMap, ok := v.(map[interface{}]interface{}); ok {
				if leftMap, ok := existing.(map[interface{}]interface{}); ok {
					merged = true
					MergeMaps(leftMap, rightMap)
				}
			}
		}

		if !merged {
			left[k] = v
		}
	}
}

func GetServices(urls []string) ([]string, error) {
	result := []string{}

	for _, url := range urls {
		indexUrl := fmt.Sprintf("%s/index.yml", url)
		content, err := LoadResource(indexUrl, true, []string{})
		if err != nil {
			log.Errorf("Failed to load %s: %v", indexUrl, err)
			continue
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			log.Errorf("Failed to unmarshal %s: %v", indexUrl, err)
			continue
		}

		if list, ok := services["services"]; ok {
			result = append(result, list...)
		}
	}

	return []string{}, nil
}

func LoadResource(location string, network bool, urls []string) ([]byte, error) {
	var bytes []byte
	err := ErrNotFound

	if strings.HasPrefix(location, "http:/") || strings.HasPrefix(location, "https:/") {
		if !network {
			return nil, ErrNoNetwork
		}
		resp, err := http.Get(location)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 http response: %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	} else if strings.HasPrefix(location, "/") {
		return ioutil.ReadFile(location)
	} else if len(location) > 0 {
		for _, url := range urls {
			ymlUrl := fmt.Sprintf("%s/%s/%s.yml", url, location[0:1], location)
			log.Infof("Loading %s from %s", location, ymlUrl)
			bytes, err = LoadResource(ymlUrl, network, []string{})
			if err == nil {
				return bytes, nil
			}
		}
	}

	return nil, err
}

func GetValue(kvPairs []string, key string) string {
	if kvPairs == nil {
		return ""
	}

	prefix := key + "="
	for _, i := range kvPairs {
		if strings.HasPrefix(i, prefix) {
			return strings.TrimPrefix(i, prefix)
		}
	}

	return ""
}
