import pytest
import rancherostest.util as u
import subprocess


ssh_command = ['./scripts/ssh', '--qemu']


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--no-format'])


@pytest.mark.timeout(40)
def test_ros_install_on_formatted_disk(qemu):
    assert qemu is not None
    u.wait_for_ssh(ssh_command)
    subprocess.check_call(ssh_command + ['sudo', 'mkfs.ext4', '/dev/vda'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.check_call(ssh_command + ['sudo', 'ros', 'install', '-f', '--no-reboot', '-d', '/dev/vda',
                                         '-i', 'rancher/os:v0.4.1'],
                          stderr=subprocess.STDOUT, universal_newlines=True)

    subprocess.call(ssh_command + ['sudo', 'reboot'],
                    stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(ssh_command)
