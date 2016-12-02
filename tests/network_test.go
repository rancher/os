package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestNetwork(c *C) {
	netArgs := []string{"-net", "nic,vlan=0,model=virtio"}
	args := []string{"--cloud-config", "./tests/assets/test_09/cloud-config.yml"}
	for i := 0; i < 7; i++ {
		args = append(args, netArgs...)
	}
	s.RunQemu(c, args...)

	s.CheckCall(c, `
cat > test-merge << "SCRIPT"
set -x -e

ip link show dev br0
ip link show dev br0.100 | grep br0.100@br0
ip link show dev eth1.100 | grep 'master br0'
ip link show dev eth6 | grep 'master bond0'
ip link show dev eth7 | grep 'master bond0'
[ "$(</sys/class/net/bond0/bonding/mode)" = "active-backup 1" ]

SCRIPT
sudo bash test-merge`)

	s.CheckCall(c, `
cat > test-merge << "SCRIPT"
set -x -e

cat /etc/resolv.conf | grep "search mydomain.com example.com"
cat /etc/resolv.conf | grep "nameserver 208.67.222.123"
cat /etc/resolv.conf | grep "nameserver 208.67.220.123"

SCRIPT
sudo bash test-merge`)
}
