// +build linux

package init

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/config"
	"github.com/rancher/os/selinux"
	"io/ioutil"
)

func initializeSelinux(c *config.CloudConfig) (*config.CloudConfig, error) {
	ret, _ := selinux.InitializeSelinux()

	if ret != 0 {
		log.Debug("Unable to initialize SELinux")
		return c, nil
	}

	// Set allow_execstack boolean to true
	if err := ioutil.WriteFile("/sys/fs/selinux/booleans/allow_execstack", []byte("1"), 0644); err != nil {
		log.Debug(err)
		return c, nil
	}

	if err := ioutil.WriteFile("/sys/fs/selinux/commit_pending_bools", []byte("1"), 0644); err != nil {
		log.Debug(err)
		return c, nil
	}

	return c, nil
}
