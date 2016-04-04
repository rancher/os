import pytest
import rostest.util as u
from rostest.util import SSH


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request)
    u.flush_out(q.stdout)
    return q


def test_ssh_key_merging(qemu):
    SSH(qemu).check_call('''cat > test-merge << "SCRIPT"
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
sudo bash test-merge
    '''.strip())
