import pytest
import rostest.util as u
from rostest.util import SSH


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request)
    u.flush_out(q.stdout)
    return q


def test_shared_mount(qemu):
    SSH(qemu).check_call('''
set -x -e

sudo mkdir /mnt/shared
sudo touch /test
sudo system-docker run --privileged -v /mnt:/mnt:shared -v /test:/test {busybox_image} mount --bind / /mnt/shared
ls /mnt/shared | grep test
    '''.format(busybox_image=u.busybox_image))
