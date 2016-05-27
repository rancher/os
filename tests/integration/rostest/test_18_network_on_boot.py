import pytest
import rostest.util as u
from rostest.util import SSH

pytestmark = pytest.mark.skipif(u.arch != 'amd64', reason='amd64 network setup impossible to replicate for arm64')

cloud_config_path = './tests/integration/assets/test_18/cloud-config.yml'

net_args_arch = {'amd64': ['-net', 'nic,vlan=1,model=virtio'],
                 'arm64': ['-device', 'virtio-net-device']}
net_args_arch['arm'] = net_args_arch['arm64']
net_args = net_args_arch[u.arch]


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request,
                   run_args=['--cloud-config', cloud_config_path] +
                   net_args + net_args + net_args)
    u.flush_out(q.stdout)
    return q


def test_network_resources_loaded(qemu):
    SSH(qemu).check_call("apt-get --version")
    SSH(qemu).check_call("sudo system-docker images | grep tianon/true")
