package main

import (
	"fmt"
	"github.com/gbazil/telnet"
)

func main() {
	t, err := telnet.Dial("192.168.1.2:23")
	if err != nil {
		fmt.Println(err)
		return
	}

	t.Read("login: ")
	t.Write("admin\n")

	t.Read("password: ")
	t.Write("qwerty\n")

	t.Read("$ ")
	t.Write("ls -l /home\n")

	s, _ := t.Read("$ ")
	fmt.Println(s)

	t.Write("exit\n")
}
