package control

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	machineUtil "github.com/docker/machine/utils"
)

func tlsConfCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "create",
			Usage:  "use it to create a new set of tls configuration certs and keys or upload existing ones",
			Action: tlsConfCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cakey",
					Usage: "path to existing certificate authority key (only use with --generate)",
				},
				cli.StringFlag{
					Name:  "ca",
					Usage: "path to existing certificate authority (only use with --genreate)",
				},
				cli.BoolFlag{
					Name:  "generate, g",
					Usage: "generate the client key and client cert from existing ca and cakey",
				},
				cli.StringFlag{
					Name:  "outDir, o",
					Usage: "the output directory to save the generated certs or keys",
				},
			},
		},
	}
}

func tlsConfCreate(c *cli.Context) {
	name := "rancher"
	bits := 2048

	caCertPath := "ca-cert.pem"
	caKeyPath := "ca-key.pem"
	outDir := "/etc/docker/tls/"
	generateCaCerts := true

	inputCaKey := ""
	inputCaCert := ""

	if val := c.String("outDir"); val != "" {
		outDir = val
	}

	if c.Bool("generate") {
		generateCaCerts = false
	}

	if val := c.String("cakey"); val != "" {
		inputCaKey = val
	}

	if val := c.String("ca"); val != "" {
		inputCaCert = val
	}

	caCertPath = filepath.Join(outDir, caCertPath)
	caKeyPath = filepath.Join(outDir, caKeyPath)

	serverCertPath := "server-cert.pem"
	serverKeyPath := "server-key.pem"

	if generateCaCerts {
		if err := machineUtil.GenerateCACertificate(caCertPath, caKeyPath, name, bits); err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		if inputCaKey == "" || inputCaCert == "" {
			fmt.Println("Path to caKey and caCert not specified with -g, searching in default location")
			inputCaKey = "/etc/docker/tls/ca-key.pem"
			inputCaCert = "/etc/docker/tls/ca-cert.pem"
		}

		if _, err := os.Stat(inputCaKey); err != nil {
			fmt.Printf("ERROR: %s does not exist\n", inputCaKey)
			return
		} else {
			caKeyPath = inputCaKey
		}

		if _, err := os.Stat(inputCaCert); err != nil {
			fmt.Printf("ERROR: %s does not exist\n", inputCaCert)
			return
		} else {
			caCertPath = inputCaCert
		}
		serverCertPath = "client-cert.pem"
		serverKeyPath = "client-key.pem"
	}

	serverCertPath = filepath.Join(outDir, serverCertPath)
	serverKeyPath = filepath.Join(outDir, serverKeyPath)

	if err := machineUtil.GenerateCert([]string{""}, serverCertPath, serverKeyPath, caCertPath, caKeyPath, name, bits); err != nil {
		fmt.Println(err.Error())
		return
	}
}
