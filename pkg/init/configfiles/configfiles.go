package configfiles

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
)

var (
	ConfigFiles = map[string][]byte{}
)

func ReadConfigFiles(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	filesToCopy := []string{
		config.CloudConfigInitFile,
		config.CloudConfigScriptFile,
		config.CloudConfigBootFile,
		config.CloudConfigNetworkFile,
		config.MetaDataFile,
		config.EtcResolvConfFile,
	}
	// And all the files in /var/log/boot/
	// TODO: I wonder if we can put this code into the log module, and have things write to the buffer until we FsReady()
	bootLog := "/var/log/"
	if files, err := ioutil.ReadDir(bootLog); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(bootLog, file.Name())
				filesToCopy = append(filesToCopy, filePath)
				log.Debugf("Swizzle: Found %s to save", filePath)
			}
		}
	}
	bootLog = "/var/log/boot/"
	if files, err := ioutil.ReadDir(bootLog); err == nil {
		for _, file := range files {
			filePath := filepath.Join(bootLog, file.Name())
			filesToCopy = append(filesToCopy, filePath)
			log.Debugf("Swizzle: Found %s to save", filePath)
		}
	}
	for _, name := range filesToCopy {
		if _, err := os.Lstat(name); !os.IsNotExist(err) {
			content, err := ioutil.ReadFile(name)
			if err != nil {
				log.Errorf("read cfg file (%s) %s", name, err)
				continue
			}
			log.Debugf("Swizzle: Saved %s to memory", name)
			ConfigFiles[name] = content
		}
	}
	return cfg, nil
}

func WriteConfigFiles(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	for name, content := range ConfigFiles {
		dirMode := os.ModeDir | 0755
		fileMode := os.FileMode(0444)
		if strings.HasPrefix(name, "/var/lib/rancher/conf/") {
			// only make the conf files harder to get to
			dirMode = os.ModeDir | 0700
			if name == config.CloudConfigScriptFile {
				fileMode = os.FileMode(0755)
			} else {
				fileMode = os.FileMode(0400)
			}
		}
		if err := os.MkdirAll(filepath.Dir(name), dirMode); err != nil {
			log.Error(err)
		}
		if err := util.WriteFileAtomic(name, content, fileMode); err != nil {
			log.Error(err)
		}
		log.Infof("Swizzle: Wrote file to %s", name)
	}
	if err := os.MkdirAll(config.VarRancherDir, os.ModeDir|0755); err != nil {
		log.Error(err)
	}
	if err := os.Chmod(config.VarRancherDir, os.ModeDir|0755); err != nil {
		log.Error(err)
	}
	log.FsReady()
	log.Debugf("WARNING: switchroot and mount OEM2 phases not written to log file")

	return cfg, nil
}
