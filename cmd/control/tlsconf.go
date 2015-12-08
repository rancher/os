package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	machineUtil "github.com/docker/machine/utils"
	"github.com/rancher/os/config"
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

		cfg, err = cfg.Merge(map[interface{}]interface{}{
			"rancher": map[interface{}]interface{}{
				"docker": map[interface{}]interface{}{
					"server_cert": string(cert),
					"server_key":  string(key),
				},
			},
		})
		if err != nil {
			return err
		}

		return cfg.Save() // certPath, keyPath are already written to by machineUtil.GenerateCert()
	}

	if err := ioutil.WriteFile(certPath, []byte(cfg.Rancher.Docker.ServerCert), 0400); err != nil {
		return err
	}

	return ioutil.WriteFile(keyPath, []byte(cfg.Rancher.Docker.ServerKey), 0400)

}

func writeCaCerts(cfg *config.CloudConfig, caCertPath, caKeyPath string) (*config.CloudConfig, error) {
	if cfg.Rancher.Docker.CACert == "" {
		if err := machineUtil.GenerateCACertificate(caCertPath, caKeyPath, NAME, BITS); err != nil {
			return nil, err
		}

		caCert, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			return nil, err
		}

		caKey, err := ioutil.ReadFile(caKeyPath)
		if err != nil {
			return nil, err
		}

		cfg, err = cfg.Merge(map[interface{}]interface{}{
			"rancher": map[interface{}]interface{}{
				"docker": map[interface{}]interface{}{
					"ca_key":  string(caKey),
					"ca_cert": string(caCert),
				},
			},
		})
		if err != nil {
			return nil, err
		}

		if err = cfg.Save(); err != nil {
			return nil, err
		}

		return cfg, nil // caCertPath, caKeyPath are already written to by machineUtil.GenerateCACertificate()
	}

	if err := ioutil.WriteFile(caCertPath, []byte(cfg.Rancher.Docker.CACert), 0400); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(caKeyPath, []byte(cfg.Rancher.Docker.CAKey), 0400); err != nil {
		return nil, err
	}

	return cfg, nil
}

func tlsConfCreate(c *cli.Context) {
	err := generate(c)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(c *cli.Context) error {
	generateServer := c.Bool("server")
	outDir := c.String("dir")
	hostnames := c.StringSlice("hostname")

	return Generate(generateServer, outDir, hostnames)
}

func Generate(generateServer bool, outDir string, hostnames []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

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

	cfg, err = writeCaCerts(cfg, caCertPath, caKeyPath)
	if err != nil {
		return err
	}

	return writeCerts(generateServer, hostnames, cfg, certPath, keyPath, caCertPath, caKeyPath)
}
