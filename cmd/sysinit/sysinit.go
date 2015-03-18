package sysinit

import (
	log "github.com/Sirupsen/logrus"
	initPkg "github.com/rancherio/os/init"
)

func Main() {
	if err := initPkg.SysInit(); err != nil {
		log.Fatal(err)
	}
}
