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
        p.terminate()

    request.addfinalizer(fin)
    return p


def rancheros_version():
    with open('./scripts/version') as f:
        for ln in iter(f.readline, ''):
            (k, _, v) = ln.partition('=')
            if k == 'VERSION' and v.strip() != '':
                return v.strip()
    raise RuntimeError("Could not parse RancherOS version")


@pytest.mark.timeout(30)
def test_system_boot(qemu):
    with qemu.stdout as f:
        for ln in iter(f.readline, ''):
            ros_booted_substr = str.find(ln, 'RancherOS {v} started'.format(v=rancheros_version()))
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
        with ssh.stdout as f:
            for ln in iter(f.readline, ''):
                print(str.strip(ln))
        ssh.wait()
        assert ssh.returncode == 0
    finally:
        if ssh.returncode is None:
            ssh.terminate()
