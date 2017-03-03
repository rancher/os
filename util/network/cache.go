package network

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func locationHash(location string) string {
	sum := md5.Sum([]byte(location))
	return hex.EncodeToString(sum[:])
}

func CacheLookup(location string) ([]byte, error) {
	cacheFile := filepath.Join(config.CacheDirectory, location)
	bytes, err := ioutil.ReadFile(cacheFile)
	if err == nil {
		log.Debugf("Using cached file: %s", cacheFile)
		return bytes, nil
	}
	log.Debugf("Cached file not found: %s", cacheFile)
	return nil, err
}

func cacheAdd(location string, data []byte) error {
	os.MkdirAll(config.CacheDirectory, 0755)
	tempFile, err := ioutil.TempFile(config.CacheDirectory, "")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(data)
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(config.CacheDirectory, location)
	cacheDir := filepath.Dir(cacheFile)
	log.Debugf("writing %s to %s", cacheFile, cacheDir)
	os.MkdirAll(cacheDir, 0755)
	os.Rename(tempFile.Name(), cacheFile)
	return nil
}
