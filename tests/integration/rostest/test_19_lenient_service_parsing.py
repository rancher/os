import time
import pytest
import rostest.util as u
from rostest.util import SSH

cloud_config_path = './tests/integration/assets/test_19/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


def test_lenient_service_parsing(qemu):
    time.sleep(5)
    SSH(qemu).check_call("sudo system-docker ps -a | grep test-parsing")
