import time

import pytest
import rostest.util as u
from rostest.util import SSH

ssh_command = ['./scripts/ssh', '--qemu', '--key', './tests/integration/assets/test.key']
cloud_config_path = './tests/integration/assets/test_03/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, ['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


nginx = {'amd64': 'nginx', 'arm': 'armhfbuild/nginx', 'arm64': 'armhfbuild/nginx'}


@pytest.mark.timeout(40)
def test_reboot_with_container_running(qemu):
    try:
        SSH(qemu, ssh_command).check_call('''
set -ex
docker run -d --restart=always %(image)s
sudo reboot
    '''.strip() % {'image': nginx[u.arch]})
    except:
        pass

    time.sleep(3)

    v = SSH(qemu, ssh_command).check_output('''
docker ps -f status=running
    '''.strip())

    assert v.find('nginx') != -1
