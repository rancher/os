import subprocess

import pytest
import rostest.util as u

ssh_command = ['./scripts/ssh', '--qemu']


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--no-format'])
    u.flush_out(q.stdout)
    return q


@pytest.mark.timeout(40)
def test_ros_install_on_formatted_disk(qemu):
    u.wait_for_ssh(qemu, ssh_command)
    subprocess.check_call(ssh_command + ['sudo', 'mkfs.ext4', '/dev/vda'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'ros', 'install', '-f', '--no-reboot', '-d', '/dev/vda',
                                         '-i', 'rancher/os:v0.4.1'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.call(ssh_command + ['sudo', 'reboot'],
                    stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(qemu, ssh_command)
