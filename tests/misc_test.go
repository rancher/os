package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestMisc(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_01/cloud-config.yml")

	s.CheckCall(c, "sudo ros env printenv FLANNEL_NETWORK | grep '10.244.0.0/16'")

	s.CheckCall(c, "ps -ef | grep 'dhcpcd -M'")

	s.CheckCall(c, `
set -e -x
sudo ros tls gen --server -H localhost
sudo ros tls gen
sudo ros c set rancher.docker.tls true
sudo system-docker restart docker
sleep 5
docker --tlsverify version`)

	s.CheckCall(c, `
set -e -x
for i in $(pidof system-docker); do
    if [ $i = 1 ]; then
        found=true
    fi
done
[ "$found" = "true" ]`)
}
