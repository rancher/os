package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestOem(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--second-drive")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -x
set -e
sudo mkfs.ext4 -L RANCHER_OEM /dev/vdb
sudo mount /dev/vdb /mnt
cat > /tmp/oem-config.yml << EOF
#cloud-config
rancher:
  upgrade:
    url: 'foo'
EOF
sudo cp /tmp/oem-config.yml /mnt
sudo umount /mnt`)

	s.Reboot(c)

	s.CheckCall(c, `
set -x
set -e
if [ ! -e /usr/share/ros/oem/oem-config.yml ]; then
    echo Failed to find /usr/share/ros/oem/oem-config.yml
    exit 1
fi

FOO="$(sudo ros config get rancher.upgrade.url)"
if [ "$FOO" != "foo" ]; then
    echo rancher.upgrade.url is not foo
    exit 1
fi`)
}
