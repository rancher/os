// +build dnspatch

package network

import "net"

func updateDNSCache() {
	net.UpdateDnsConf()
}
