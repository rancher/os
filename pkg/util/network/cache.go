package network

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"

	"github.com/burmilla/os/pkg/log"
)

const (
	cacheDirectory = "/var/lib/rancher/cache/"
)

func locationHash(location string) string {
	sum := md5.Sum([]byte(location))
	return hex.EncodeToString(sum[:])
}

func cacheLookup(location string) []byte {
	cacheFile := cacheDirectory + locationHash(location)
	bytes, err := ioutil.ReadFile(cacheFile)
	if err == nil {
		log.Debugf("Using cached file: %s", cacheFile)
		return bytes
	}
	return nil
}

func cacheAdd(location string, data []byte) {
	tempFile, err := ioutil.TempFile(cacheDirectory, "")
	if err != nil {
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(data)
	if err != nil {
		return
	}

	cacheFile := cacheDirectory + locationHash(location)
	os.Rename(tempFile.Name(), cacheFile)
}

func cacheMove(location string) (string, error) {
	cacheFile := cacheDirectory + locationHash(location)
	tempFile := cacheFile + "_temp"
	if err := os.Rename(cacheFile, tempFile); err != nil {
		return "", err
	}
	return tempFile, nil
}

func cacheMoveBack(name string) error {
	return os.Rename(name, strings.TrimRight(name, "_temp"))
}
