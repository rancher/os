package control

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/os/log"
)

func yes(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(line[0:1]) == "y"
}
