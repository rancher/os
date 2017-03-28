package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestNetworkFromUrl(c *C) {
	netArgs := []string{"-net", "nic,vlan=0,model=virtio"}
	args := []string{"--append", "rancher.debug=true rancher.password=test-me rancher.cloud_init.datasources=[url:https://gist.githubusercontent.com/joshwget/0bdc616cd26162ad87c535644c8b1ef6/raw/8cce947c08cf006e932b71d92ddbb96bae8e3325/gistfile1.txt]"}
	for i := 0; i < 7; i++ {
		args = append(args, netArgs...)
	}
	s.RunQemuWithNetConsole(c, args...)

	s.NetCheckCall(c, `
cat > test-merge << "SCRIPT"
set -x -e

ip link show dev br0
ip link show dev br0.100 | grep br0.100@br0
ip link show dev eth1.100 | grep 'master br0'

SCRIPT
sudo bash test-merge`)

	s.NetCheckCall(c, `
cat > test-merge << "SCRIPT"
set -x -e

cat /etc/resolv.conf | grep "search mydomain.com example.com"
cat /etc/resolv.conf | grep "nameserver 208.67.222.123"
cat /etc/resolv.conf | grep "nameserver 208.67.220.123"

SCRIPT
sudo bash test-merge`)
}

func (s *QemuSuite) TestNoNetworkCloudConfigFromUrl(c *C) {
	args := []string{
		"--no-network",
		"--append",
		"rancher.debug=true rancher.password=test-me rancher.cloud_init.datasources=[url:https://gist.githubusercontent.com/joshwget/0bdc616cd26162ad87c535644c8b1ef6/raw/8cce947c08cf006e932b71d92ddbb96bae8e3325/gistfile1.txt]",
	}
	s.RunQemuWithNetConsole(c, args...)

	s.NetCheckCall(c, "sudo ros config get rancher.log | grep true")
}

func (s *QemuSuite) TestNoNetworkConsoleSwitch(c *C) {
	args := []string{
		"--no-network",
		"--append",
		"rancher.debug=true rancher.password=test-me rancher.console=alpine",
	}
	s.RunQemuWithNetConsole(c, args...)

	s.NetCheckCall(c, "uname -a")
}
