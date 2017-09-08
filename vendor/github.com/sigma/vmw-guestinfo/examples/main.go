package main

import (
	"fmt"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

func main() {
	if !vmcheck.IsVirtualWorld() {
		fmt.Println("not in a virtual world... :(")
		return
	}

	config := rpcvmx.NewConfig()

	fmt.Println(config.SetString("foo", "bar"))
	fmt.Println(config.String("foo", "foo"))

	fmt.Println(config.SetInt("foo", 3))
	fmt.Println(config.Int("foo", 0))

	fmt.Println(config.SetBool("foo", false))
	fmt.Println(config.Bool("foo", true))

}
