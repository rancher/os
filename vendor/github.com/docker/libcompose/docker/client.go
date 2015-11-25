package docker

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/docker/pkg/tlsconfig"
	"github.com/samalba/dockerclient"
)

const (
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

func init() {
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.ConfigDir()
	}
}

type ClientOpts struct {
	TLS        bool
	TLSVerify  bool
	TLSOptions tlsconfig.Options
	TrustKey   string
	Host       string
}

func CreateClient(c ClientOpts) (dockerclient.Client, error) {
	if c.TLSOptions.CAFile == "" {
		c.TLSOptions.CAFile = filepath.Join(dockerCertPath, defaultCaFile)
	}
	if c.TLSOptions.CertFile == "" {
		c.TLSOptions.CertFile = filepath.Join(dockerCertPath, defaultCertFile)
	}
	if c.TLSOptions.KeyFile == "" {
		c.TLSOptions.KeyFile = filepath.Join(dockerCertPath, defaultKeyFile)
	}

	if c.Host == "" {
		defaultHost := os.Getenv("DOCKER_HOST")
		if defaultHost == "" {
			if runtime.GOOS != "windows" {
				// If we do not have a host, default to unix socket
				defaultHost = fmt.Sprintf("unix://%s", opts.DefaultUnixSocket)
			} else {
				// If we do not have a host, default to TCP socket on Windows
				defaultHost = fmt.Sprintf("tcp://%s:%d", opts.DefaultHTTPHost, opts.DefaultHTTPPort)
			}
		}
		defaultHost, err := opts.ValidateHost(defaultHost)
		if err != nil {
			return nil, err
		}
		c.Host = defaultHost
	}

	if c.TrustKey == "" {
		c.TrustKey = filepath.Join(homedir.Get(), ".docker", defaultTrustKeyFile)
	}

	if c.TLSVerify {
		c.TLS = true
	}

	if c.TLS {
		c.TLSOptions.InsecureSkipVerify = !c.TLSVerify
	}

	var tlsConfig *tls.Config

	if c.TLS {
		var err error
		tlsConfig, err = tlsconfig.Client(c.TLSOptions)
		if err != nil {
			return nil, err
		}
	}

	return dockerclient.NewDockerClient(c.Host, tlsConfig)
}
