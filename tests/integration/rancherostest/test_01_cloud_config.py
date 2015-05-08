import pytest
import rancherostest.util as u
import subprocess


ssh_command = ['ssh', '-p', '2222', '-F', './assets/scripts_ssh_config', '-i', './tests/integration/assets/test.key',
               'rancher@localhost']


@pytest.fixture(scope="module")
def qemu(request):
    return u.run_qemu(request, ['--cloud-config', './tests/integration/assets/cloud-config-1.yml'])


@pytest.mark.timeout(40)
def test_check_whats_going_on(qemu):
    assert qemu.returncode is None
    u.wait_for_ssh(ssh_command)

    ssh = subprocess.Popen(
        ssh_command + ['sudo', 'rancherctl', 'c', 'get', 'environment'],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    with ssh, ssh.stdout as f:
        for ln in iter(f.readline, ''):
            print(str.strip(ln))

    assert ssh.returncode == 0
