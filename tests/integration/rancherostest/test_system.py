import pytest
import string
import subprocess


@pytest.fixture(scope="module")
def qemu(request):
    p = subprocess.Popen('./scripts/run', stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

    def fin():
        p.stdout.close()
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(20)
def test_system_boot(qemu):
    for ln in iter(qemu.stdout.readline, ''):
        l = string.strip(ln)
        ros_booted_substr = string.find(l, 'RancherOS v0.3.1-rc2 started')
        if ros_booted_substr > -1:
            assert True
            return
