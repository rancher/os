package control

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	machineUtil "github.com/docker/machine/utils"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

const (
	NAME string = "rancher"
	BITS int    = 2048
)

func tlsConfCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "generate",
			ShortName: "gen",
			Usage:     "generates new set of TLS configuration certs",
			Action:    tlsConfCreate,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "hostname, H",
					Usage: "the hostname for which you want to generate the certificate",
					Value: &cli.StringSlice{"localhost"},
				},
				cli.BoolFlag{
					Name:  "server, s",
					Usage: "generate the server keys instead of client keys",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "the directory to save/read the certs to/from",
					Value: "",
				},
			},
		},
	}
}

func writeCerts(generateServer bool, hostname []string, cfg *config.CloudConfig, certPath, keyPath, caCertPath, caKeyPath string) error {
	if !generateServer {
		return machineUtil.GenerateCert([]string{""}, certPath, keyPath, caCertPath, caKeyPath, NAME, BITS)
	}

	if cfg.Rancher.Docker.ServerKey == "" || cfg.Rancher.Docker.ServerCert == "" {
		err := machineUtil.GenerateCert(hostname, certPath, keyPath, caCertPath, caKeyPath, NAME, BITS)
		if err != nil {
			return err
		}

		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return err
		}

		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return err
		}

		// certPath, keyPath are already written to by machineUtil.GenerateCert()
		if err := config.Set("rancher.docker.server_cert", string(cert)); err != nil {
			return err
		}
		if err := config.Set("rancher.docker.server_key", string(key)); err != nil {
			return err
		}
	}

	cfg = config.LoadConfig()

	if err := util.WriteFileAtomic(certPath, []byte(cfg.Rancher.Docker.ServerCert), 0400); err != nil {
		return err
	}

	return util.WriteFileAtomic(keyPath, []byte(cfg.Rancher.Docker.ServerKey), 0400)

}

func writeCaCerts(cfg *config.CloudConfig, caCertPath, caKeyPath string) error {
	if cfg.Rancher.Docker.CACert == "" {
		if err := machineUtil.GenerateCACertificate(caCertPath, caKeyPath, NAME, BITS); err != nil {
			return err
		}

		caCert, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			return err
		}

		caKey, err := ioutil.ReadFile(caKeyPath)
		if err != nil {
			return err
		}

		// caCertPath, caKeyPath are already written to by machineUtil.GenerateCACertificate()
		if err := config.Set("rancher.docker.ca_cert", string(caCert)); err != nil {
			return err
		}
		if err := config.Set("rancher.docker.ca_key", string(caKey)); err != nil {
			return err
		}
	}

	cfg = config.LoadConfig()

	if err := util.WriteFileAtomic(caCertPath, []byte(cfg.Rancher.Docker.CACert), 0400); err != nil {
		return err
	}

	if err := util.WriteFileAtomic(caKeyPath, []byte(cfg.Rancher.Docker.CAKey), 0400); err != nil {
		return err
	}

	return nil
}

func tlsConfCreate(c *cli.Context) error {
	err := generate(c)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func generate(c *cli.Context) error {
	generateServer := c.Bool("server")
	outDir := c.String("dir")
	hostnames := c.StringSlice("hostname")

	return Generate(generateServer, outDir, hostnames)
}

func Generate(generateServer bool, outDir string, hostnames []string) error {
	if outDir == "" {
		if generateServer {
			outDir = "/etc/docker/tls"
		} else {
			outDir = "/home/rancher/.docker"
		}
		log.Infof("Out directory (-d, --dir) not specified, using default: %s", outDir)
	}
	caCertPath := filepath.Join(outDir, "ca.pem")
	caKeyPath := filepath.Join(outDir, "ca-key.pem")
	certPath := filepath.Join(outDir, "cert.pem")
	keyPath := filepath.Join(outDir, "key.pem")

	if generateServer {
		certPath = filepath.Join(outDir, "server-cert.pem")
		keyPath = filepath.Join(outDir, "server-key.pem")
	}

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0700); err != nil {
			return err
		}
	}

	cfg := config.LoadConfig()

	err := writeCaCerts(cfg, caCertPath, caKeyPath)
	if err != nil {
		return err
	}
	if err := writeCerts(generateServer, hostnames, cfg, certPath, keyPath, caCertPath, caKeyPath); err != nil {
		return err
	}

	if !generateServer {
		if err := filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
			return os.Chown(path, 1100, 1100) // rancher:rancher
		}); err != nil {
			return err
		}
	}

	return nil
}
