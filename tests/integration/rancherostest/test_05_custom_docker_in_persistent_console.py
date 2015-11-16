import pytest
import rancherostest.util as u
import subprocess


ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_05/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--cloud-config', cloud_config_path])


@pytest.mark.timeout(40)
def test_system_docker_survives_custom_docker_install(qemu):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)
    subprocess.check_call(ssh_command + ['curl', '-OL',
                                         'https://experimental.docker.com/builds/Linux/x86_64/docker-1.10.0-dev'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['chmod', '+x', '/home/rancher/docker-1.10.0-dev'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'ln', '-sf', '/home/rancher/docker-1.10.0-dev', '/usr/bin/docker'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'system-docker', 'restart', 'docker'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'system-docker', 'version'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(ssh_command)
