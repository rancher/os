import pytest
import subprocess
import time


@pytest.fixture(scope="module")
def qemu(request):
    p = subprocess.Popen('./scripts/run', stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    def fin():
        print('\nTerminating QEMU')
        p.stdout.close()
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(20)
def test_system_boot(qemu):
    for ln in iter(qemu.stdout.readline, ''):
        ros_booted_substr = str.find(ln, 'RancherOS v0.3.1-rc2 started')  # TODO use ./scripts/version
        print(str.strip(ln))
        if ros_booted_substr > -1:
            assert True
            return
    assert False


@pytest.mark.timeout(40)
def test_run_system_container(qemu):
    assert qemu.returncode is None
    time.sleep(2)  # or else ssh will fail (WTF?!)
    ssh = subprocess.Popen(
        './scripts/ssh sudo system-docker run --rm busybox /bin/true', shell=True,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    try:
        for ln in iter(ssh.stdout.readline, ''):
            print(str.strip(ln))
            pass
        ssh.wait()
        assert ssh.returncode == 0
    finally:
        ssh.stdout.close()
        if ssh.returncode is None:
            ssh.terminate()
