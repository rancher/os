package integration

import (
	"fmt"
	"time"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestInstallMsDosMbr(c *C) {
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

func (s *QemuSuite) TestUpgradeFromImage(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh

	//TODO: --fresh isn't giving us a new disk why? (that's what the parted print is for atm)
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("running %s", version)
		s.CheckCall(c, "sudo uname -a")
		//TODO: detect "last release, and install that
		s.CheckCall(c, `
echo "---------------------------------- generic"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml -i rancher/os:v0.7.1 --append "console=ttyS0 rancher.password=rancher"
#TODO copy installer image, new ros, and new kernel to HD, so we can fake things up next time? (or have next boot from HD, but have the iso available..)
sudo mkdir -p /bootiso
sudo mount -t iso9660 /dev/sr0 /bootiso/
sudo mount /dev/vda1 /mnt/
sudo mkdir -p /mnt/rancher-installer/build/ 
sudo cp /bootiso/rancheros/installer.tar.gz /mnt/rancher-installer/build/
sudo cp /bootiso/rancheros/Dockerfile.amd64 /mnt/rancher-installer/build/
sudo cp -r /bootiso/boot /mnt/rancher-installer/build/
sudo cp /bin/ros /mnt/rancher-installer/build/
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	{
		runArgs = []string{
			"--boothd",
		}
		s.RunQemuWith(c, runArgs...)

		s.CheckOutput(c, "ros version v0.7.1\n", Equals, "sudo ros -v")
		s.CheckCall(c, "sudo uname -a")

		// load the installer.tar.gz, get the other install files into an image, and runit.
		s.CheckCall(c, `sudo system-docker run --name builder -dt --volumes-from system-volumes -v /:/host alpine sh
	sudo system-docker exec -t builder ln -s /host/rancher-installer/build/ros /bin/system-docker
	sudo system-docker exec -t builder system-docker load -i /host/rancher-installer/build/installer.tar.gz
	sudo system-docker exec -t builder system-docker build -t qwer -f /host/rancher-installer/build/Dockerfile.amd64 /host/rancher-installer/build
	sudo ros os upgrade -i qwer --no-reboot -f --append "console=tty0 console=ttyS0 rancher.password=rancher"
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
	s.CheckCall(c, "sudo uname -a")
	s.Stop(c)
}
