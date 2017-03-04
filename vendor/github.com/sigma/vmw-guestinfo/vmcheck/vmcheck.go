package vmcheck

import (
	"github.com/sigma/bdoor"
)

// IsVirtualWorld returns whether the code is running in a VMware virtual machine or no
func IsVirtualWorld() bool {
	return bdoor.HypervisorPortCheck()
}
