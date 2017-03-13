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

func (s *QemuSuite) TestNetworkCfg(c *C) {
	args := []string{"--cloud-config", "./tests/assets/multi_nic/cloud-config.yml"}
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=0,model=virtio"}...)
	s.RunQemu(c, args...)

	// TODO: work out why the ipv6 loopback isn't present
	//    inet6 ::1/128 scope host
	//       valid_lft forever preferred_lft forever

	// show ip a output without mac addresses
	s.CheckOutput(c, `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    inet 10.1.0.41/24 scope global eth0
       valid_lft forever preferred_lft forever
3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    inet 10.31.168.85/24 scope global eth1
       valid_lft forever preferred_lft forever
4: eth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    inet6 XX::XX:XX:XX:XX/64 scope link 
       valid_lft forever preferred_lft forever
5: eth3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth3
       valid_lft forever preferred_lft forever
    inet6 XX::XX:XX:XX:XX/64 scope link 
       valid_lft forever preferred_lft forever
6: docker-sys: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000
    inet 172.18.42.2/16 scope global docker-sys
       valid_lft forever preferred_lft forever
    inet6 XX::XX:XX:XX:XX/64 scope link 
       valid_lft forever preferred_lft forever
8: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default 
    inet 172.17.0.1/16 scope global docker0
       valid_lft forever preferred_lft forever
`, Equals, "ip a | grep -v ether | sed 's/inet 10\\.0\\.2\\..*\\/24 brd/inet XX.XX.XX.XX\\/24 brd/' | sed 's/inet6 .*\\/64 scope/inet6 XX::XX:XX:XX:XX\\/64 scope/'")

	s.CheckOutput(c, `Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         10.31.168.1     0.0.0.0         UG    0      0        0 eth1
0.0.0.0         10.0.2.2        0.0.0.0         UG    205    0        0 eth3
10.0.2.0        0.0.0.0         255.255.255.0   U     205    0        0 eth3
10.1.0.0        0.0.0.0         255.255.255.0   U     0      0        0 eth0
10.31.168.0     0.0.0.0         255.255.255.0   U     0      0        0 eth1
172.17.0.0      0.0.0.0         255.255.0.0     U     0      0        0 docker0
172.18.0.0      0.0.0.0         255.255.0.0     U     0      0        0 docker-sys
`, Equals, "route -n")
}
