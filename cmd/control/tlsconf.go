package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	machineUtil "github.com/docker/machine/utils"
	"github.com/rancherio/os/config"
)

const (
	NAME string = "rancher"
	BITS int    = 2048
)

func tlsConfCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "generate",
			Usage:  "generates new set of TLS configuration certs",
			Action: tlsConfCreate,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "hostname",
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

	if cfg.Rancher.UserDocker.ServerKey == "" || cfg.Rancher.UserDocker.ServerCert == "" {
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

		return cfg.SetConfig(&config.CloudConfig{
			Rancher: config.RancherConfig{
				UserDocker: config.DockerConfig{
					CAKey:      cfg.Rancher.UserDocker.CAKey,
					CACert:     cfg.Rancher.UserDocker.CACert,
					ServerCert: string(cert),
					ServerKey:  string(key),
				},
			},
		})
	}

	if err := ioutil.WriteFile(certPath, []byte(cfg.Rancher.UserDocker.ServerCert), 0400); err != nil {
		return err
	}

	return ioutil.WriteFile(keyPath, []byte(cfg.Rancher.UserDocker.ServerKey), 0400)

}

func writeCaCerts(cfg *config.CloudConfig, caCertPath, caKeyPath string) error {
	if cfg.Rancher.UserDocker.CACert == "" {
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

		err = cfg.SetConfig(&config.CloudConfig{
			Rancher: config.RancherConfig{
				UserDocker: config.DockerConfig{
					CAKey:  string(caKey),
					CACert: string(caCert),
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	}

	if err := ioutil.WriteFile(caCertPath, []byte(cfg.Rancher.UserDocker.CACert), 0400); err != nil {
		return err
	}

	return ioutil.WriteFile(caKeyPath, []byte(cfg.Rancher.UserDocker.CAKey), 0400)
}

func tlsConfCreate(c *cli.Context) {
	err := generate(c)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(c *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	generateServer := c.Bool("server")
	outDir := c.String("dir")
	if outDir == "" {
		return fmt.Errorf("out directory (-d, --dir) not specified")
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

	if err := writeCaCerts(cfg, caCertPath, caKeyPath); err != nil {
		return err
	}

	hostnames := c.StringSlice("hostname")
	return writeCerts(generateServer, hostnames, cfg, certPath, keyPath, caCertPath, caKeyPath)
}
