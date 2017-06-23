package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestReadDatasourcesFromDisk(c *C) {
	s.RunQemu(c)

	s.CheckCall(c, `
sudo tee /var/lib/rancher/conf/cloud-config.d/datasources.yml << EOF
rancher:
  cloud_init:
    datasources:
    - url:https://gist.githubusercontent.com/joshwget/e1c49f8b1ddeeba01bc9d0a3be01ed60/raw/9168b380fde182d53acea487d49b680648a0ca5b/gistfile1.txt
EOF
`)

	s.Reboot(c)

	s.CheckCall(c, "sudo ros config get rancher.log | grep true")
}

func (s *QemuSuite) TestIsoCloudConfig(c *C) {
	runArgs := []string{
		"--fresh",
		"--cloud-config-iso",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckCall(c, `ls .ssh/authorized_keys`)
}

func (s *QemuSuite) TestFatCloudConfig(c *C) {
	runArgs := []string{
		"--fresh",
		"--cloud-config-fat",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckCall(c, `ls .ssh/authorized_keys`)
}
