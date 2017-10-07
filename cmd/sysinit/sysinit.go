package sysinit

import (
	initPkg "github.com/rancher/os/init"
	"github.com/rancher/os/log"
	"io/ioutil"
	"os"
)

func Main() {
	log.InitLogger()

	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Infof("2Resolv.conf == [%s], %s", resolve, err)
	log.Infof("Exec %v", os.Args)

	if err := initPkg.SysInit(); err != nil {
		log.Fatal(err)
	}
}
