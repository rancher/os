package control

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/os/config"
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

func formatImage(image string, cfg *config.CloudConfig) string {
	domainRegistry := cfg.Rancher.Environment["REGISTRY_DOMAIN"]
	if domainRegistry != "docker.io" && domainRegistry != "" {
		return fmt.Sprintf("%s/%s", domainRegistry, image)
	}
	return image
}
