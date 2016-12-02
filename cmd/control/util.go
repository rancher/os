package control

import (
	"fmt"
	"strings"

	"github.com/rancher/os/log"
)

func yes(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	var line string
	_, err := fmt.Scan(&line)
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(line[0:1]) == "y"
}
