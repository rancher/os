import pytest
import rancherostest.util as u
import subprocess
import yaml


ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_03/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--cloud-config', cloud_config_path])


@pytest.fixture(scope="module")
def cloud_config():
    return yaml.load(open(cloud_config_path))


@pytest.mark.timeout(40)
def test_reboot_with_container_running(qemu):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)
    subprocess.check_call(ssh_command + ['docker', 'run', '-d', '--restart=always', 'nginx'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.call(ssh_command + ['sudo', 'reboot'],
                    stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(ssh_command)
    v = subprocess.check_output(ssh_command + ['docker', 'ps', '-f', 'status=running'],
                                stderr=subprocess.STDOUT, universal_newlines=True)

    assert v.find('nginx') != -1
