package network

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"

	"github.com/rancher/os/pkg/log"
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

func cacheRemove(location string) error {
	cacheFile := cacheDirectory + locationHash(location)
	return os.Remove(cacheFile)
}
