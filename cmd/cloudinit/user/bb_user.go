package user

import (
	"os/exec"
	"strings"
)

func bb_addUser(name, gecos, homedir, passwd string, noCreateHome, system bool) error {
	userAddArgs := []string{}
	if gecos != "" {
		userAddArgs = append(userAddArgs, "-g ", gecos)
	}
	userAddArgs = append(userAddArgs, "-D")
	if !noCreateHome && homedir != "" {
		userAddArgs = append(userAddArgs, "-h ", homedir)
	} else {
		userAddArgs = append(userAddArgs, "-H")
	}
	if system {
		userAddArgs = append(userAddArgs, "-S")
	}
	userAddArgs = append(userAddArgs, name)
	cmd := exec.Command("adduser", userAddArgs...)
	if err := cmd.Run(); err != nil {
		return err
	}
	if passwd == "" {
		return nil
	}
	return bb_chPasswd(name, passwd)
}

func bb_chPasswd(user, passwd string) error {
	cmd := exec.Command("chpasswd", "-e")
	cmd.Stdin = strings.NewReader(user + ":" + passwd)
	return cmd.Run()
}
