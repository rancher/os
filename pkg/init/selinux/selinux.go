//go:build linux
// +build linux

package selinux

import (
	"io/ioutil"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/selinux"
)

func Initialize(c *config.CloudConfig) (*config.CloudConfig, error) {
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
