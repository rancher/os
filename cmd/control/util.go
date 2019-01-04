package control

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"
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

func symLinkEngineBinary(version string) []symlink {
	baseSymlink := []symlink{
		{"/var/lib/rancher/engine/docker", "/usr/bin/docker"},
		{"/var/lib/rancher/engine/dockerd", "/usr/bin/dockerd"},
		{"/var/lib/rancher/engine/docker-init", "/usr/bin/docker-init"},
		{"/var/lib/rancher/engine/docker-proxy", "/usr/bin/docker-proxy"},
		{"/usr/share/ros/os-release", "/usr/lib/os-release"},
		{"/usr/share/ros/os-release", "/etc/os-release"},
	}
	if strings.Contains(version, "18.09") {
		baseSymlink = append(baseSymlink, []symlink{
			{"/var/lib/rancher/engine/containerd", "/usr/bin/containerd"},
			{"/var/lib/rancher/engine/ctr", "/usr/bin/ctr"},
			{"/var/lib/rancher/engine/containerd-shim", "/usr/bin/containerd-shim"},
			{"/var/lib/rancher/engine/runc", "/usr/bin/runc"},
		}...)
	} else {
		baseSymlink = append(baseSymlink, []symlink{
			{"/var/lib/rancher/engine/docker-containerd", "/usr/bin/docker-containerd"},
			{"/var/lib/rancher/engine/docker-containerd-ctr", "/usr/bin/docker-containerd-ctr"},
			{"/var/lib/rancher/engine/docker-containerd-shim", "/usr/bin/docker-containerd-shim"},
			{"/var/lib/rancher/engine/docker-runc", "/usr/bin/docker-runc"},
		}...)
	}
	return baseSymlink
}
