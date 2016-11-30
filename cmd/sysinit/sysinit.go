package sysinit

import (
	initPkg "github.com/rancher/os/init"
	"github.com/rancher/os/log"
)

func Main() {
	log.InitLogger()
	if err := initPkg.SysInit(); err != nil {
		log.Fatal(err)
	}
}
