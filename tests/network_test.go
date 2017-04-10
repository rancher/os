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

func (s *QemuSuite) TestNetworkBootCfg(c *C) {
	args := []string{"--append", "rancher.network.interfaces.eth1.address=10.1.0.41/24 rancher.network.interfaces.eth1.gateway=10.1.0.1 rancher.network.interfaces.eth0.dhcp=true"}
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=0,model=virtio"}...)
	s.RunQemu(c, args...)
	s.CheckOutput(c,
		"1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1\n"+
			"    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00\n"+
			"    inet 127.0.0.1/8 scope XXXX lo\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 ::1/128 scope host \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth0\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3456/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet 10.1.0.41/24 scope global eth1\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"4: eth2: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000\n"+
			"5: eth3: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000\n"+
			"6: docker-sys: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000\n"+
			"    inet 172.18.42.2/16 scope global docker-sys\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 XX::XX:XX:XX:XX/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"8: docker0: XXXXXXX......\n"+
			"    inet 172.17.0.1/16 scope global docker0\n"+
			"       valid_lft forever preferred_lft forever\n",
		Equals,
		"ip a | "+
			"grep -v ether | "+
			// TODO: figure out why sometimes loopback is scope global
			"sed 's/scope host lo/scope XXXX lo/g' | sed 's/scope global lo/scope XXXX lo/g' | "+
			"sed 's/inet 10\\.0\\.2\\..*\\/24 brd/inet XX.XX.XX.XX\\/24 brd/' | "+
			"sed 's/8: docker0: .*/8: docker0: XXXXXXX....../g' | "+
			"sed '/inet6 fe80::5054:ff:fe12:.*\\/64/!s/inet6 .*\\/64 scope/inet6 XX::XX:XX:XX:XX\\/64 scope/'",
		// fe80::18b6:9ff:fef5:be33
	)
}

func (s *QemuSuite) TestNetworkBootAndCloudCfg(c *C) {
	args := []string{
		"--append", "rancher.network.interfaces.eth1.address=10.1.0.52/24 rancher.network.interfaces.eth1.gateway=10.1.0.1 rancher.network.interfaces.eth0.dhcp=true rancher.network.interfaces.eth3.dhcp=true",
		"--cloud-config", "./tests/assets/multi_nic/cloud-config.yml",
	}
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=1,model=virtio"}...)
	args = append(args, []string{"-net", "nic,vlan=0,model=virtio"}...)
	s.RunQemu(c, args...)
	s.CheckOutput(c,
		"1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1\n"+
			"    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00\n"+
			"    inet 127.0.0.1/8 scope XXXX lo\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 ::1/128 scope host \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth0\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3456/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			// This shows that the boot cmdline wins over the cloud-config
			// But IIRC, the cloud-init metadata wins allowing you to use ip4ll to get the hoster's metadata
			// Need a test for that (presumably once we have libmachine based tests)
			"    inet 10.1.0.52/24 scope global eth1\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"4: eth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet 10.31.168.85/24 scope global eth2\n"+
			"       valid_lft forever preferred_lft forever\n"+
			// TODO: I think it would be better if this was dhcp: false, but it could go either way
			"5: eth3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth3\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3459/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"6: docker-sys: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000\n"+
			"    inet 172.18.42.2/16 scope global docker-sys\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 XX::XX:XX:XX:XX/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"8: docker0: XXXXXXX......\n"+
			"    inet 172.17.0.1/16 scope global docker0\n"+
			"       valid_lft forever preferred_lft forever\n",
		Equals,
		"ip a | "+
			"grep -v ether | "+
			// TODO: figure out why sometimes loopback is scope global
			"sed 's/scope host lo/scope XXXX lo/g' | sed 's/scope global lo/scope XXXX lo/g' | "+
			"sed 's/inet 10\\.0\\.2\\..*\\/24 brd/inet XX.XX.XX.XX\\/24 brd/' | "+
			"sed 's/8: docker0: .*/8: docker0: XXXXXXX....../g' | "+
			"sed '/inet6 fe80::5054:ff:fe12:.*\\/64/!s/inet6 .*\\/64 scope/inet6 XX::XX:XX:XX:XX\\/64 scope/'",
		// fe80::18b6:9ff:fef5:be33
	)
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
	s.CheckOutput(c,
		"1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1\n"+
			"    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00\n"+
			"    inet 127.0.0.1/8 scope XXXX lo\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 ::1/128 scope host \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth0\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3456/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet 10.1.0.41/24 scope global eth1\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"4: eth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet 10.31.168.85/24 scope global eth2\n"+
			"       valid_lft forever preferred_lft forever\n"+
			// TODO: I think it would be better if this was dhcp: false, but it could go either way
			"5: eth3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth3\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3459/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"6: docker-sys: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000\n"+
			"    inet 172.18.42.2/16 scope global docker-sys\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 XX::XX:XX:XX:XX/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n"+
			"8: docker0: XXXXXXX......\n"+
			"    inet 172.17.0.1/16 scope global docker0\n"+
			"       valid_lft forever preferred_lft forever\n",
		Equals,
		"ip a | "+
			"grep -v ether | "+
			// TODO: figure out why sometimes loopback is scope global
			"sed 's/scope host lo/scope XXXX lo/g' | sed 's/scope global lo/scope XXXX lo/g' | "+
			"sed 's/inet 10\\.0\\.2\\..*\\/24 brd/inet XX.XX.XX.XX\\/24 brd/' | "+
			"sed 's/8: docker0: .*/8: docker0: XXXXXXX....../g' | "+
			"sed '/inet6 fe80::5054:ff:fe12:.*\\/64/!s/inet6 .*\\/64 scope/inet6 XX::XX:XX:XX:XX\\/64 scope/'",
		// fe80::18b6:9ff:fef5:be33
	)

	s.CheckOutput(c,
		"Kernel IP routing table\n"+
			"Destination     Gateway         Genmask         Flags Metric Ref    Use Iface\n"+
			"0.0.0.0         10.1.0.1        0.0.0.0         UG    0      0        0 eth1\n"+
			"0.0.0.0         10.0.2.2        0.0.0.0         UG    202    0        0 eth0\n"+
			"0.0.0.0         10.0.2.2        0.0.0.0         UG    205    0        0 eth3\n"+
			"10.0.2.0        0.0.0.0         255.255.255.0   U     202    0        0 eth0\n"+
			"10.0.2.0        0.0.0.0         255.255.255.0   U     205    0        0 eth3\n"+
			"10.1.0.0        0.0.0.0         255.255.255.0   U     0      0        0 eth1\n"+
			"10.31.168.0     0.0.0.0         255.255.255.0   U     0      0        0 eth2\n"+
			"172.17.0.0      0.0.0.0         255.255.0.0     U     0      0        0 docker0\n"+
			"172.18.0.0      0.0.0.0         255.255.0.0     U     0      0        0 docker-sys\n",
		Equals, "route -n")

	s.CheckCall(c, "sudo ros config set rancher.network.interfaces.eth3.dhcp true")
	//s.CheckCall(c, "sudo netconf")
	s.Reboot(c)
	s.CheckOutput(c,
		"5: eth3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000\n"+
			"    inet XX.XX.XX.XX/24 brd 10.0.2.255 scope global eth3\n"+
			"       valid_lft forever preferred_lft forever\n"+
			"    inet6 fe80::5054:ff:fe12:3459/64 scope link \n"+
			"       valid_lft forever preferred_lft forever\n",
		Equals,
		"ip a show eth3 | "+
			"grep -v ether | "+
			// TODO: figure out why sometimes loopback is scope global
			"sed 's/scope host lo/scope XXXX lo/g' | sed 's/scope global lo/scope XXXX lo/g' | "+
			"sed 's/inet 10\\.0\\.2\\..*\\/24 brd/inet XX.XX.XX.XX\\/24 brd/' | "+
			"sed '/inet6 fe80::5054:ff:fe12:.*\\/64/!s/inet6 .*\\/64 scope/inet6 XX::XX:XX:XX:XX\\/64 scope/'",
	)
}
