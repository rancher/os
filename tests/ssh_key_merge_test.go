package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestSshKeyMerge(c *C) {
	s.RunQemu(c)
	s.CheckCall(c, `
cat > test-merge << "SCRIPT"
set -x -e
rm /var/lib/rancher/conf/cloud-config.yml

EXISTING=$(ros config get ssh_authorized_keys | head -1)
cat > /var/lib/rancher/conf/metadata << EOF
SSHPublicKeys:
  "0": zero
  "1": one
  "2": two
EOF
ros config set hostname one
ros config set hostname two
ros config set hostname three

cat > expected << EOF
$EXISTING
- zero
- one
- two

EOF

ros config get ssh_authorized_keys > got

diff got expected

SCRIPT
sudo bash test-merge`)
}
