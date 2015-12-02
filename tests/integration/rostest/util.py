import itertools as it
import pytest
import subprocess
import time


ros_test = 'ros-test'


def iter_lines(s):
    return it.imap(str.rstrip, iter(s.readline, ''))


def strip_comment(prefix):
    return lambda s: s.partition(prefix)[0].strip()


def non_empty(s):
    return s != ''


def parse_value(var):
    def get_value(s):
        (k, _, v) = s.partition('=')
        (k, v) = (k.strip(), v.strip())
        if k == var:
            return v
        return ''
    return get_value


def with_effect(p):
    def effect(s):
        p(s)
        return s
    return effect


def rancheros_version(build_conf):
    with open(build_conf) as f:
        for v in it.ifilter(non_empty,
                            it.imap(parse_value('VERSION'),
                                    it.ifilter(non_empty,
                                               it.imap(strip_comment('#'), iter_lines(f))))):
            return v
    raise RuntimeError("Could not parse RancherOS version")


def run_qemu(request, run_args=[]):
    print('\nStarting QEMU')
    p = subprocess.Popen(['./scripts/run', '--qemu', '--no-rebuild', '--no-rm-usr', '--fresh'] + run_args,
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    def fin():
        print('\nTerminating QEMU')
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(10)
def wait_for_ssh(ssh_command=['./scripts/ssh', '--qemu']):
    i = 0
    print('\nWaiting for ssh and docker... ' + str(i))
    while subprocess.call(ssh_command + ['docker version >/dev/null 2>&1']) != 0:
        i += 1
        print('\nWaiting for ssh and docker... ' + str(i))
        time.sleep(1)
