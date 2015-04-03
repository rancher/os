package hostname

import (
	"github.com/rancherio/os/util"
)

func SetHostname(hostname string) error {
	osType := util.GetOSType()

	if osType == "busybox" {
		return bb_setHostname(hostname)
	}
	return bb_setHostname(hostname)
}
