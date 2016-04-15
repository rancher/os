import pytest
import rostest.util as u
from rostest.util import SSH

ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_12/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


@pytest.mark.timeout(40)
def test_dhcp_hostname(qemu):
    SSH(qemu, ssh_command).check_call('''
hostname | grep rancher-dev
cat /etc/hosts | grep rancher-dev
    ''')
