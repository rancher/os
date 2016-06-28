package glue

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var resolvConf = "/etc/resolv.conf"

func SetupResolvConf(state *DockerPluginState) error {
	root := state.Spec.Root.Path
	mode := state.HostConfig.NetworkMode
	targetFile := path.Join(root, resolvConf)

	if !isZero(targetFile) {
		return nil
	}

	if mode.IsHost() || mode.IsNone() {
		return copyToExistingFile(targetFile, resolvConf)
	}

	if mode.IsContainer() {
		return nil
	}

	f, err := os.OpenFile(resolvConf, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	ignoreNameserver := false
	buf := &bytes.Buffer{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		copyLine := true
		switch {
		case strings.HasPrefix(line, "nameserver"):
			copyLine = false
			if ignoreNameserver {
				break
			}
			for _, dns := range state.HostConfig.DNS {
				ignoreNameserver = true
				buf.WriteString(fmt.Sprintf("nameserver %s\n", dns))
			}
			if !ignoreNameserver && strings.Contains(line, "127.0.") {
				ignoreNameserver = true
				buf.WriteString(fmt.Sprintf("nameserver 8.8.8.8\n"))
				buf.WriteString(fmt.Sprintf("nameserver 8.8.4.4\n"))
			} else {
				copyLine = true
			}
		case strings.HasPrefix(line, "search"):
			if len(state.HostConfig.DNSSearch) > 0 {
				buf.WriteString(fmt.Sprintf("search %s\n", strings.Join(state.HostConfig.DNSSearch, " ")))
			}
		case strings.HasPrefix(line, "options"):
			if len(state.HostConfig.DNSOptions) > 0 {
				buf.WriteString(fmt.Sprintf("options %s\n", strings.Join(state.HostConfig.DNSOptions, " ")))
			}
		}

		if copyLine {
			buf.WriteString(line)
			buf.WriteRune('\n')
		}
	}

	if err := s.Err(); err != nil {
		return err
	}

	return ioutil.WriteFile(targetFile, buf.Bytes(), 0666)
}
