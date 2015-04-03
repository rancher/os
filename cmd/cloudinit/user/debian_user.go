package user

import (
	"os/exec"
	"strings"
)

func deb_addUser(name, gecos, homedir, passwd string, noCreateHome, system bool) error {
	userAddArgs := []string{}
	if gecos != "" {
		userAddArgs = append(userAddArgs, "--gecos", gecos)
	}
	userAddArgs = append(userAddArgs, "--disabled-password")
	if !noCreateHome && homedir != "" {
		userAddArgs = append(userAddArgs, "--home", homedir)
	} else {
		userAddArgs = append(userAddArgs, "--no-create-home")
	}
	if system {
		userAddArgs = append(userAddArgs, "--system")
	}
	userAddArgs = append(userAddArgs, name)
	cmd := exec.Command("adduser", userAddArgs...)
	if err := cmd.Run(); err != nil {
		return err
	}
	if passwd == "" {
		return nil
	}
	return deb_chPasswd(name, passwd)
}

func deb_chPasswd(user, passwd string) error {
	cmd := exec.Command("chpasswd", "-e")
	cmd.Stdin = strings.NewReader(user + ":" + passwd)
	return cmd.Run()
}
