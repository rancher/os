package hostname

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/rancher/os/config"
)

func SetHostnameFromCloudConfig(cc *config.CloudConfig) error {
	var hostname string
	if cc.Hostname == "" {
		hostname = cc.Rancher.Defaults.Hostname
	} else {
		hostname = cc.Hostname
	}

	if hostname == "" {
		return nil
	}

	// set hostname
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	return nil
}

func SyncHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if hostname == "" {
		return nil
	}

	hosts, err := os.Open("/etc/hosts")
	defer hosts.Close()
	if err != nil {
		return err
	}
	lines := bufio.NewScanner(hosts)
	hostsContent := ""
	for lines.Scan() {
		line := strings.TrimSpace(lines.Text())
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "127.0.1.1" {
			hostsContent += "127.0.1.1 " + hostname + "\n"
			continue
		}
		hostsContent += line + "\n"
	}
	if err := ioutil.WriteFile("/etc/hosts", []byte(hostsContent), 0600); err != nil {
		return err
	}

	return nil
}
