package sysinit

import (
	log "github.com/Sirupsen/logrus"
	initPkg "github.com/rancher/os/init"
)

func Main() {
	if err := initPkg.SysInit(); err != nil {
		log.Fatal(err)
	}
}
