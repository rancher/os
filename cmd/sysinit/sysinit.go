package sysinit

import (
	"io/ioutil"
	"os"

	initPkg "github.com/rancher/os/init"
	"github.com/rancher/os/log"
)

func Main() {
	log.InitLogger()

	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Infof("Resolv.conf == [%s], %v", resolve, err)
	log.Infof("Exec %v", os.Args)

	if err := initPkg.SysInit(); err != nil {
		log.Fatal(err)
	}
}
