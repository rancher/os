from __future__ import print_function

import subprocess

import pytest
import rostest.util as u


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request)


@pytest.mark.timeout(30)
def test_system_boot(qemu):
    u.flush_out(qemu.stdout)


busybox = {'amd64': 'busybox', 'arm': 'armhf/busybox', 'arm64': 'aarch64/busybox'}


@pytest.mark.timeout(60)
def test_run_system_container(qemu):
    u.wait_for_ssh(qemu)

    ssh = subprocess.Popen(
        './scripts/ssh --qemu sudo system-docker run --rm ' + busybox[u.arch] + ' /bin/true',
        shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    for ln in u.iter_lines(ssh.stdout):
        print(ln)
    ssh.wait()

    assert ssh.returncode == 0


@pytest.mark.timeout(60)
def test_ros_dev(qemu):
    u.wait_for_ssh(qemu)

    ssh = subprocess.Popen(
        './scripts/ssh --qemu sudo ros dev',
        shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    for ln in u.iter_lines(ssh.stdout):
        print(ln)
    ssh.wait()

    assert ssh.returncode == 0
