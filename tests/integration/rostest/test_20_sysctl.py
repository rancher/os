import pytest
import rostest.util as u
from rostest.util import SSH

cloud_config_path = './tests/integration/assets/test_20/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


def test_sysctl(qemu):
    SSH(qemu).check_call("sudo cat /proc/sys/kernel/domainname | grep test")
    SSH(qemu).check_call("sudo cat /proc/sys/dev/cdrom/debug | grep 1")
