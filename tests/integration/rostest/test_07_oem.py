import time
import pytest
import rostest.util as u
from rostest.util import SSH


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--append', 'rancher.state.dev=x'])
    u.flush_out(q.stdout)
    return q


def test_oem(qemu):
    try:
        SSH(qemu).check_call('''
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
sudo umount /mnt
sudo reboot
        '''.strip())
    except:
        pass

    time.sleep(3)

    SSH(qemu).check_call('''
set -x
if [ ! -e /usr/share/ros/oem/oem-config.yml ]; then
    echo Failed to find /usr/share/ros/oem/oem-config.yml
    exit 1
fi

FOO="$(sudo ros config get rancher.upgrade.url)"
if [ "$FOO" != "foo" ]; then
    echo rancher.upgrade.url is not foo
    exit 1
fi
    '''.strip())
