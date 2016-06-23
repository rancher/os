import pytest
import rostest.util as u
from rostest.util import SSH

cloud_config_path = './tests/integration/assets/test_21/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--second-drive', '--cloud-config',
                   cloud_config_path])
    u.flush_out(q.stdout)
    return q


def test_swap(qemu):
    SSH(qemu).check_call("sudo mkswap /dev/vdb")
    SSH(qemu).check_call("sudo cloud-init -execute")
    SSH(qemu).check_call("cat /proc/swaps | grep /dev/vdb")
