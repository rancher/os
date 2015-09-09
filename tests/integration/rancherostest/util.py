import itertools as it
import pytest
import subprocess
import time


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


def run_qemu(request, run_args=[]):
    subprocess.check_call('rm -f ./state/empty-hd.img', shell=True)
    print('\nrm ./state/*')
    print('\nStarting QEMU')
    p = subprocess.Popen(['./scripts/run'] + run_args,
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    def fin():
        print('\nTerminating QEMU')
        p.terminate()

    request.addfinalizer(fin)
    return p


@pytest.mark.timeout(10)
def wait_for_ssh(ssh_command=['./scripts/ssh']):
    while subprocess.call(ssh_command + ['/bin/true']) != 0:
        time.sleep(1)
