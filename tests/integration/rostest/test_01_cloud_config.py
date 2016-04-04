import string

import pytest
import rostest.util as u
from rostest.util import SSH
import yaml

ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_01/cloud-config.yml'


net_args = {'amd64': ['-net', 'nic,vlan=1,model=virtio,macaddr=52:54:00:12:34:59',
                      '-net', 'user,vlan=1,net=10.10.2.0/24'],
            'arm64': ['-netdev', 'user,id=net1,net=10.10.2.0/24',
                      '-device', 'virtio-net-device,netdev=net1,mac=52:54:00:12:34:59']}
net_args['arm'] = net_args['arm64']


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--cloud-config', cloud_config_path] + net_args[u.arch])
    u.flush_out(q.stdout)
    return q


@pytest.fixture(scope="module")
def cloud_config():
    return yaml.load(open(cloud_config_path))


@pytest.mark.timeout(40)
def test_ssh_authorized_keys(qemu):
    u.wait_for_ssh(qemu, ssh_command)
    assert True


@pytest.mark.timeout(40)
def test_rancher_environment(qemu, cloud_config):
    v = SSH(qemu, ssh_command).check_output('''
sudo ros env printenv FLANNEL_NETWORK
    '''.strip())

    assert v.strip() == cloud_config['rancher']['environment']['FLANNEL_NETWORK']


@pytest.mark.timeout(40)
def test_docker_args(qemu, cloud_config):
    v = SSH(qemu, ssh_command).check_output('''
ps -ef | grep docker
    '''.strip())

    expected = string.join(cloud_config['rancher']['docker']['args'])

    assert v.find(expected) != -1


@pytest.mark.timeout(40)
def test_dhcpcd(qemu, cloud_config):
    v = SSH(qemu, ssh_command).check_output('''
ps -ef | grep dhcpcd
    '''.strip())

    assert v.find('dhcpcd -M') != -1


@pytest.mark.timeout(40)
def test_services_include(qemu, cloud_config):
    u.wait_for_ssh(qemu, ssh_command, ['docker inspect kernel-headers >/dev/null 2>&1'])


@pytest.mark.timeout(40)
def test_docker_tls_args(qemu, cloud_config):
    SSH(qemu, ssh_command).check_call('''
set -e -x
sudo ros tls gen
sleep 5
docker --tlsverify version
    '''.strip())


@pytest.mark.timeout(40)
def test_rancher_network(qemu, cloud_config):
    v = SSH(qemu, ssh_command).check_output('''
ip route get to 10.10.2.120
    '''.strip())

    assert v.split(' ')[5] + '/24' == \
        cloud_config['rancher']['network']['interfaces']['mac=52:54:00:12:34:59']['address']


def test_docker_not_pid_one(qemu):
    SSH(qemu, ssh_command).check_call('''
set -e -x
for i in $(pidof docker); do
    [ $i != 1 ]
done
    '''.strip())
