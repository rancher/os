import os
import string
import subprocess
import pytest


@pytest.mark.timeout(20)
def test_system_boot():
    os.chdir('../..')
    p = subprocess.Popen('./scripts/run', stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    try:
        for ln in iter(p.stdout.readline, ''):
            l = string.strip(ln)
            ros_booted_substr = string.find(l, 'RancherOS v0.3.1-rc2 started')
            if ros_booted_substr > -1:
                assert True
                return
    finally:
        p.stdout.close()
        p.terminate()
        p.wait()
