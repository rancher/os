from __future__ import print_function
import itertools as it
import pytest
import subprocess
import rostest.util as u


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request)


@pytest.mark.timeout(30)
def test_system_boot(qemu):
    version = u.rancheros_version('./build.conf')
    print('parsed version: ' + version)

    def has_ros_started_substr(s):
        return str.find(s, 'RancherOS {v} started'.format(v=version)) > -1

    for _ in it.ifilter(has_ros_started_substr,
                        it.imap(u.with_effect(print), u.iter_lines(qemu.stdout))):
        assert True
        return
    assert False


@pytest.mark.timeout(60)
def test_run_system_container(qemu):
    assert qemu is not None
    u.wait_for_ssh()

    ssh = subprocess.Popen(
        './scripts/ssh --qemu sudo system-docker run --rm busybox /bin/true',
        shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    for ln in u.iter_lines(ssh.stdout):
        print(ln)
    ssh.wait()

    assert ssh.returncode == 0
