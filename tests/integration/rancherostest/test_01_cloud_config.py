import pytest
import rancherostest.util as u
import subprocess
import yaml


ssh_command = ['ssh', '-p', '2222', '-F', './assets/scripts_ssh_config', '-i', './tests/integration/assets/test.key',
               'rancher@localhost']
cloud_config_path = './tests/integration/assets/cloud-config-01.yml'


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--cloud-config', cloud_config_path,
                                '-net', 'nic,vlan=1,model=virtio', '-net', 'user,vlan=1,net=10.10.2.0/24'])


@pytest.fixture(scope="module")
def cloud_config():
    return yaml.load(open(cloud_config_path))


@pytest.mark.timeout(40)
def test_ssh_authorized_keys(qemu):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)
    assert True


@pytest.mark.timeout(40)
def test_rancher_environment(qemu, cloud_config):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)

    v = subprocess.check_output(
        ssh_command + ['sudo', 'rancherctl', 'env', 'printenv', 'FLANNEL_NETWORK'],
        stderr=subprocess.STDOUT, universal_newlines=True)

    assert v.strip() == cloud_config['rancher']['environment']['FLANNEL_NETWORK']


@pytest.mark.timeout(40)
def test_rancher_network(qemu, cloud_config):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)

    v = subprocess.check_output(
        ssh_command + ['ip', 'route', 'get', 'to', '10.10.2.120'],
        stderr=subprocess.STDOUT, universal_newlines=True)

    assert v.split(' ')[2] == 'eth1'
    assert v.split(' ')[5] + '/24' == cloud_config['rancher']['network']['interfaces']['eth1']['address']
