import pytest
import rostest.util as u
from rostest.util import SSH


cloud_config_path = './tests/integration/assets/test_09/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--cloud-config', cloud_config_path,
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio',
                                      '-net', 'nic,vlan=0,model=virtio'])
    u.flush_out(q.stdout)
    return q


def test_network_conf(qemu):
    SSH(qemu).check_call('bash', '-c', '''cat > test-merge << "SCRIPT"
set -x -e

ip link show dev br0
ip link show dev br0.100 | grep br0.100@br0
ip link show dev eth1.100 | grep 'master br0'
ip link show dev eth6 | grep 'master bond0'
ip link show dev eth7 | grep 'master bond0'
[ "$(</sys/class/net/bond0/bonding/mode)" = "active-backup 1" ]

SCRIPT
sudo bash test-merge
    '''.strip())
