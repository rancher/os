import subprocess

import pytest
import rostest.util as u

ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_05/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q

docker_url = {'amd64': 'https://experimental.docker.com/builds/Linux/x86_64/docker-1.10.0-dev',
              'arm': 'https://github.com/rancher/docker/releases/download/v1.10.3-arm/docker-1.10.3_arm',
              'arm64': 'https://github.com/rancher/docker/releases/download/v1.10.3-arm/docker-1.10.3_arm64'}


@pytest.mark.timeout(40)
def test_system_docker_survives_custom_docker_install(qemu):
    u.wait_for_ssh(qemu, ssh_command)
    subprocess.check_call(ssh_command + ['curl', '-Lfo', './docker', docker_url[u.arch]],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['chmod', '+x', '/home/rancher/docker'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'ln', '-sf', '/home/rancher/docker', '/usr/bin/docker'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'system-docker', 'restart', 'docker'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'system-docker', 'version'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(qemu, ssh_command)
