package user

func AddUser(name, gecos, homedir, passwd string, noCreateHome, system bool) error {
	osType := getOsType()
	if osType == "busybox" {
		return bb_addUser(name, gecos, homedir, passwd, noCreateHome, system)
	} else if osType == "debian" {
		return deb_addUser(name, gecos, homedir, passwd, noCreateHome, system)
	} else if osType == "rhel" {
		//TBD
	} else if osType == "fedora" {
		//TBD
	}
	return bb_addUser(name, gecos, homedir, passwd, noCreateHome, system)
}

func getOsType() string {
	return "busybox"
}
