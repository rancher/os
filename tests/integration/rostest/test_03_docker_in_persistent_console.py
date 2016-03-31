import subprocess

import pytest
import rostest.util as u

ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_03/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


nginx = {'amd64': 'nginx', 'arm': 'armhfbuild/nginx', 'arm64': 'armhfbuild/nginx'}


@pytest.mark.timeout(40)
def test_reboot_with_container_running(qemu):
    u.wait_for_ssh(qemu, ssh_command)
    subprocess.check_call(ssh_command + ['docker', 'run', '-d', '--restart=always', nginx[u.arch]],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.call(ssh_command + ['sudo', 'reboot'],
                    stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(qemu, ssh_command)
    v = subprocess.check_output(ssh_command + ['docker', 'ps', '-f', 'status=running'],
                                stderr=subprocess.STDOUT, universal_newlines=True)

    assert v.find('nginx') != -1
