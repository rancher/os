package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestRosConfig(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_14/cloud-config.yml")

	s.CheckCall(c, `
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
 fi`)

	s.CheckCall(c, `
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
fi`)

	s.CheckCall(c, `
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
 fi`)

	s.CheckCall(c, `
set -x -e

if sudo ros config export | grep "PRIVATE KEY"; then
    exit 1
fi

sudo ros config export --private | grep "PRIVATE KEY"

sudo ros config export --full | grep "udev"
sudo ros config export --private --full | grep "ntp"
sudo ros config export --full | grep "labels"

sudo ros config export --private --full | grep "PRIVATE KEY"`)
}
