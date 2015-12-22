from __future__ import print_function

import itertools as it
import subprocess
import time

import pytest

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
    assert p.returncode is None

    def fin():
        print('\nTerminating QEMU')
        p.terminate()
        p.wait()

    request.addfinalizer(fin)
    return p


def has_substr(token):
    return lambda s: str.find(s, token) > -1


def flush_out(stdout, substr='RancherOS '):
    for _ in it.ifilter(has_substr(substr),
                        it.imap(with_effect(print), iter_lines(stdout))):
        return


@pytest.mark.timeout(10)
def wait_for_ssh(qemu, ssh_command=['./scripts/ssh', '--qemu'], command=['docker version >/dev/null 2>&1']):
    i = 0
    assert qemu.returncode is None
    print('\nWaiting for ssh and docker... ' + str(i))
    while subprocess.call(ssh_command + command) != 0:
        i += 1
        print('\nWaiting for ssh and docker... ' + str(i))
        time.sleep(1)
        assert qemu.returncode is None


class SSH:
    def __init__(self, qemu, ssh_command=['./scripts/ssh', '--qemu']):
        self._qemu = qemu
        self._ssh_command = ssh_command
        self._waited = False

    def check_call(self, *args, **kw):
        if not self._waited:
            wait_for_ssh(self._qemu, ssh_command=self._ssh_command)
            self._waited = True

        kw['stderr'] = subprocess.STDOUT
        kw['universal_newlines'] = True
        return subprocess.check_call(self._ssh_command + list(args), **kw)
