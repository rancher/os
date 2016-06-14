import subprocess

import pytest
import rostest.util as u
from rostest.util import SSH

ssh_command = ['./scripts/ssh', '--qemu']


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--no-format'])
    u.flush_out(q.stdout)
    return q


@pytest.mark.timeout(40)
def test_ros_install_on_formatted_disk(qemu):
    u.wait_for_ssh(qemu, ssh_command)

    subprocess.check_call(
        ['sh', '-c', 'docker save rancher/os:%s%s | ./scripts/ssh sudo system-docker load' % (u.version, u.suffix)],
        stderr=subprocess.STDOUT, universal_newlines=True)

    SSH(qemu, ssh_command).check_call('''
set -e -x
sudo mkfs.ext4 /dev/vda
sudo ros install -f --no-reboot -d /dev/vda -i rancher/os:%s%s
    '''.strip() % (u.version, u.suffix))

    subprocess.call(ssh_command + ['sudo', 'reboot'],
                    stderr=subprocess.STDOUT, universal_newlines=True)

    u.wait_for_ssh(qemu, ssh_command)
