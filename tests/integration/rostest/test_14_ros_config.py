import pytest
import rostest.util as u
from rostest.util import SSH

cloud_config_path = './tests/integration/assets/test_14/cloud-config.yml'


@pytest.fixture(scope="module")
def qemu(request):
    q = u.run_qemu(request, run_args=['--cloud-config', cloud_config_path])
    u.flush_out(q.stdout)
    return q


def test_ros_config_string(qemu):
    SSH(qemu).check_call('''
set -x -e

if [ "$(sudo ros config get hostname)" == "hostname3
 " ]; then
    sudo ros config get hostname
    exit 1
 fi

sudo ros config set hostname rancher-test
if [ "$(sudo ros config get hostname)" == "rancher-test
 " ]; then
    sudo ros config get hostname
    exit 1
 fi
    '''.strip())


def test_ros_config_bool(qemu):
    SSH(qemu).check_call('''
set -x -e

if [ "$(sudo ros config get rancher.log)" == "true
 " ]; then
    sudo ros config get rancher.log
    exit 1
 fi

sudo ros config set rancher.log false
if [ "$(sudo ros config get rancher.log)" == "false
 " ]; then
    sudo ros config get rancher.log
    exit 1
 fi

if [ "$(sudo ros config get rancher.debug)" == "false
 " ]; then
    sudo ros config get rancher.debug
    exit 1
 fi

sudo ros config set rancher.debug true
if [ "$(sudo ros config get rancher.debug)" == "true
 " ]; then
    sudo ros config get rancher.debug
    exit 1
 fi
    '''.strip())


def test_ros_config_slice(qemu):
    SSH(qemu).check_call('''
set -x -e

sudo ros config set rancher.network.dns.search '[a,b]'
if [ "$(sudo ros config get rancher.network.dns.search)" == "- a
 - b

 " ]; then
    sudo ros config get rancher.network.dns.search
    exit 1
 fi

sudo ros config set rancher.network.dns.search '[]'
if [ "$(sudo ros config get rancher.network.dns.search)" == "[]
 " ]; then
    sudo ros config get rancher.network.dns.search
    exit 1
 fi
    '''.strip())
