import pytest
import subprocess
import time


@pytest.fixture(scope="module")
def qemu(request):
    subprocess.check_call('rm ./state/*', shell=True)
    print('\nrm ./state/*')
    print('\nStarting QEMU')
    p = subprocess.Popen('./scripts/run', stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    def fin():
        print('\nTerminating QEMU')
        p.stdout.close()
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(30)
def test_system_boot(qemu):
    for ln in iter(qemu.stdout.readline, ''):
        ros_booted_substr = str.find(ln, 'RancherOS v0.3.1-rc2 started')  # TODO use ./scripts/version
        print(str.strip(ln))
        if ros_booted_substr > -1:
            assert True
            return
    assert False


@pytest.mark.timeout(10)
def wait_for_ssh():
    while subprocess.call(['./scripts/ssh', '/bin/true']) != 0:
        time.sleep(1)


@pytest.mark.timeout(40)
def test_run_system_container(qemu):
    assert qemu.returncode is None
    wait_for_ssh()
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
