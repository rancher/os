from __future__ import print_function
import itertools as it
import pytest
import subprocess
import rostest.util as u


@pytest.fixture(scope="module")
def build_and_run(request):
    print('\nBuilding and running RancherOS with custom kernel')
    p = subprocess.Popen(['./scripts/docker-run.sh', '--rm',
                          './tests/integration/assets/test_02/test-custom-kernel.sh'],
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    def fin():
        print('\nTerminating docker-run test-custom-kernel')
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(30)
def test_system_boot(build_and_run):
    version = u.rancheros_version('./tests/integration/assets/test_02/build.conf')
    print('parsed version: ' + version)

    def has_ros_started_substr(s):
        return str.find(s, 'RancherOS {v} started'.format(v=version)) > -1

    for _ in it.ifilter(has_ros_started_substr,
                        it.imap(u.with_effect(print), u.iter_lines(build_and_run.stdout))):
        assert True
        return
    assert False
