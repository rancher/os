package sysinit

import (
	"io/ioutil"
	"os"

	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/sysinit"
)

func Main() {
	log.InitLogger()

	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Infof("Resolv.conf == [%s], %v", resolve, err)
	log.Infof("Exec %v", os.Args)

	if err := sysinit.SysInit(); err != nil {
		log.Fatal(err)
	}
}
