import pytest
import rostest.util as u
from rostest.util import SSH


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--append', 'rancher.state.directory=ros_subdir'])
    u.flush_out(q.stdout)
    return q


def test_system_docker_survives_custom_docker_install(qemu):
    SSH(qemu).check_call('''
set -x -e
mkdir x
sudo mount $(sudo ros dev LABEL=RANCHER_STATE) x
[ -d x/ros_subdir/home/rancher ]
    '''.strip())
