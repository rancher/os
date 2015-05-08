import pytest
import rancherostest.util as u
import subprocess
import yaml


ssh_command = ['ssh', '-p', '2222', '-F', './assets/scripts_ssh_config', '-i', './tests/integration/assets/test.key',
               'rancher@localhost']
cloud_config_path = './tests/integration/assets/cloud-config-01.yml'


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--cloud-config', cloud_config_path])


@pytest.fixture(scope="module")
def cloud_config_01():
    return yaml.load(open(cloud_config_path))


@pytest.mark.timeout(40)
def test_ssh_authorized_keys(qemu):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)
    assert True


@pytest.mark.timeout(40)
def test_rancher_environment(qemu, cloud_config_01):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)

    v = subprocess.check_output(
        ssh_command + ['sudo', 'rancherctl', 'env', 'printenv', 'FLANNEL_NETWORK'],
        stderr=subprocess.STDOUT, universal_newlines=True)

    assert v.strip() == cloud_config_01['rancher']['environment']['FLANNEL_NETWORK']
