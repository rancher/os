package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestOem(c *C) {
	err := s.RunQemu("--append", "rancher.state.dev=x")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -x
set -e
sudo mkfs.ext4 -L RANCHER_OEM /dev/vda
sudo mount /dev/vda /mnt
cat > /tmp/oem-config.yml << EOF
#cloud-config
rancher:
  upgrade:
    url: 'foo'
EOF
sudo cp /tmp/oem-config.yml /mnt
sudo umount /mnt`)

	s.Reboot()

	s.CheckCall(c, `
set -x
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
