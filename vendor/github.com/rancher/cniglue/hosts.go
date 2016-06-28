package glue

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/pkg/errors"
)

var (
	hostsFile    = "/etc/hosts"
	defaultHosts = []byte(`127.0.0.1	localhost
::1	localhost ip6-localhost ip6-loopback
fe00::0	ip6-localnet
ff00::0	ip6-mcastprefix
ff02::1	ip6-allnodes
ff02::2	ip6-allrouters
`)
)

func SetupHosts(state *DockerPluginState, cniResult *types.Result) error {
	targetFile := path.Join(state.Spec.Root.Path, hostsFile)
	mode := state.HostConfig.NetworkMode

	if !isZero(targetFile) {
		return nil
	}

	if mode.IsHost() {
		return copyToExistingFile(targetFile, hostsFile)
	} else if mode.IsNone() {
		return writeHosts(targetFile, "", "")
	} else if mode.IsContainer() {
		return nil
	}

	ip := ""
	if cniResult != nil && cniResult.IP4 != nil {
		ip = cniResult.IP4.IP.String()
	}

	return writeHosts(targetFile, ip, state.Config.Hostname)
}

func writeHosts(file, ip, hostname string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Wrap(err, "opening "+file)
	}
	defer f.Close()

	if _, err := f.Write(defaultHosts); err != nil {
		return err
	}

	if ip != "" && hostname != "" {
		_, err := io.WriteString(f, fmt.Sprintf("%s %s\n", ip, hostname))
		if err != nil {
			return err
		}
	}

	return nil
}
