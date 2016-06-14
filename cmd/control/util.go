package control

import (
	"bufio"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func yes(in *bufio.Reader, question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(line[0:1]) == "y"
}
