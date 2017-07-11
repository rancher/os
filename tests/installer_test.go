package integration

import (
	"fmt"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestInstallMsDosMbr(c *C) {
	// test_17 cloud config is an invalid http proxy cfg, so the installer has no network
	runArgs := []string{
		"--iso",
		"--fresh",
		"--cloud-config",
		"./tests/assets/test_17/cloud-config.yml",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)
		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
echo "---------------------------------- generic"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append rancher.password=rancher
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.Stop(c)
}

func (s *QemuSuite) TestInstallGptMbr(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
echo "---------------------------------- gptsyslinux"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -t gptsyslinux -c config.yml
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	// TEST parted output? (gpt non-uefi == legacy_boot)
	s.Stop(c)
}

func (s *QemuSuite) TestInstallAlpine(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		s.MakeCall("sudo ros console switch -f alpine")
		c.Assert(s.WaitForSSH(), IsNil)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
set -ex
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.Stop(c)
}

func (s *QemuSuite) TestAutoResize(c *C) {
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	disk := "/dev/vda1\n"
	size := ""
	{
		s.RunQemuWith(c, runArgs...)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
set -ex
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append "rancher.resize_device=/dev/vda"
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.CheckCall(c, "sudo mount "+strings.TrimSpace(disk)+" /mnt")
		size = s.CheckOutput(c, size, Not(Equals), "df -h | grep "+strings.TrimSpace(disk)+" | head -n1 | sed 's/ \\+/;/g' | cut -d ';' -f 2")
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
		"--resizehd", "+20G",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.CheckOutput(c, disk, Equals, "blkid | cut -f 1 -d ' ' | sed 's/://'")
	s.CheckOutput(c, size, Not(Equals), "df -h | grep "+strings.TrimSpace(disk)+" | head -n1 | sed 's/ \\+/;/g' | cut -d ';' -f 2")

	s.Stop(c)
}

func (s *QemuSuite) KillsMyServerTestInstalledDhcp(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
		//		"-net", "nic,vlan=0,model=virtio",
		//		"-net", "user,vlan=0",
		//		"-net", "nic,vlan=0,model=virtio",
		//		"-net", "user,vlan=0",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		s.MakeCall("ip a")

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
echo "rancher:" >> config.yml
echo "  network:" >> config.yml
echo "    interfaces:" >> config.yml
echo "      eth2:" >> config.yml
echo "        dhcp: true" >> config.yml
echo "      eth1:" >> config.yml
echo "        address: 10.0.2.253/24" >> config.yml
echo "        dhcp: false" >> config.yml
echo "        gateway: 10.0.2.1" >> config.yml
echo "        mtu: 1500" >> config.yml
ip a
echo "==================="
cat config.yml | sudo ros config merge
sudo ros service stop network
sleep 1
sudo ros service start network
sleep 1
ip a
echo "==================="
sudo ros install --force --no-reboot --device /dev/vda -c config.yml -a "console=ttyS0 rancher.autologin=ttyS0 console=ttyS1 rancher.autologin=ttyS1 rancher.debug=true"
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	runArgs = []string{
		"--boothd",
		"-net", "nic,vlan=0,model=virtio",
		"-net", "user,vlan=0",
		"-net", "nic,vlan=0,model=virtio",
		"-net", "user,vlan=0",
	}
	s.RunQemuWithNetConsole(c, runArgs...)

	s.NetCheckOutput(c, version, Equals, "sudo ros -v")
	s.NetCheckOutput(c, "", Not(Equals), "sh", "-c", "ip a show eth1 | grep 10.0.2..253")
	s.Stop(c)
}

func (s *QemuSuite) TestConfigDhcp(c *C) {
	runArgs := []string{
		"--iso",
		"--fresh",
		"-net", "nic,vlan=0,model=virtio",
		"-net", "user,vlan=0",
		"-net", "nic,vlan=0,model=virtio",
		"-net", "user,vlan=0",
	}
	version := ""
	{
		s.RunQemuWithNetConsole(c, runArgs...)

		s.NetCall("ip a")

		version = s.NetCheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.NetCheckCall(c, `
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
echo "rancher:" >> config.yml
echo "  network:" >> config.yml
echo "    interfaces:" >> config.yml
echo "      eth2:" >> config.yml
echo "        dhcp: true" >> config.yml
echo "      eth1:" >> config.yml
echo "        address: 10.0.2.253/24" >> config.yml
echo "        dhcp: false" >> config.yml
echo "        gateway: 10.0.2.1" >> config.yml
echo "        mtu: 1500" >> config.yml
ip a
echo "==================="
cat config.yml | sudo ros config merge
sudo ros service stop network
sleep 1
sudo ros service start network
sleep 1
echo "==================="
sudo system-docker logs network
echo "==================="
ip a
`)

		s.NetCheckOutput(c, version, Equals, "sudo ros -v")
		s.NetCheckOutput(c, "", Not(Equals), "sh", "-c", "\"ip a show eth1 | grep 10.0.2.253\"")
		s.Stop(c)
	}
}
