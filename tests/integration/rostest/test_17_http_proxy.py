import pytest
import rostest.util as u
from rostest.util import SSH

cloud_config_path = './tests/integration/assets/test_17/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


def test_docker_http_proxy(qemu):
    SSH(qemu).check_call('''
set -x -e

sudo system-docker exec docker env | grep HTTP_PROXY=invalid
sudo system-docker exec docker env | grep HTTPS_PROXY=invalid
sudo system-docker exec docker env | grep NO_PROXY=invalid

if docker pull busybox; then
    exit 1
else
    exit 0
fi
    ''')


def test_system_docker_http_proxy(qemu):
    try:
        SSH(qemu).check_call('sudo reboot')
    except:
        pass

    SSH(qemu).check_call('''
set -x -e

if sudo system-docker pull busybox; then
    exit 1
else
    exit 0
fi
    ''')
