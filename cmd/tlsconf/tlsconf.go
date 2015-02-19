package tlsconf

import (
	"fmt"
	"os"
	"path/filepath"

	machineUtil "github.com/docker/machine/utils"
)

func Main() {
	name := "rancher"
	bits := 2048

	vargs := os.Args

	caCertPath := "ca.pem"
	caKeyPath := "ca-key.pem"
	outDir := "/etc/docker/tls/"
	generateCaCerts := true

	inputCaKey := ""
	inputCaCert := ""

	for index := range vargs {
		arg := vargs[index]
		if arg == "--help" || arg == "-h" {
			fmt.Println("run tlsconfig with no args to generate ca, cakey, server-key and server-cert in /var/run \n")
			fmt.Println("--help or -h\t print this help text")
			fmt.Println("--cakey\t\t path to existing certificate authority key (only use with -g)")
			fmt.Println("--ca\t\t path to existing certificate authority (only use with -g)")
			fmt.Println("--g \t\t generates server key and server cert from existing ca and caKey")
			fmt.Println("--outdir \t the output directory to save the generate certs or keys")
			return
		} else if arg == "--outdir" {
			if len(vargs) > index+1 {
				outDir = vargs[index+1]
			} else {
				fmt.Println("please specify a output directory")
			}
		} else if arg == "-g" {
			generateCaCerts = false
		} else if arg == "--cakey" {
			if len(vargs) > index+1 {
				inputCaKey = vargs[index+1]
			} else {
				fmt.Println("please specify a input ca-key file path")
			}
		} else if arg == "--ca" {
			if len(vargs) > index+1 {
				inputCaCert = vargs[index+1]
			} else {
				fmt.Println("please specify a input ca file path")
			}
		}
	}

	caCertPath = filepath.Join(outDir, caCertPath)
	caKeyPath = filepath.Join(outDir, caKeyPath)

	if generateCaCerts {
		if err := machineUtil.GenerateCACertificate(caCertPath, caKeyPath, name, bits); err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		if inputCaKey == "" || inputCaCert == "" {
			fmt.Println("Please specify caKey and CaCert along with -g")
			return
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
	}

	serverCertPath := "server-cert.pem"
	serverCertPath = filepath.Join(outDir, serverCertPath)

	serverKeyPath := "server-key.pem"
	serverKeyPath = filepath.Join(outDir, serverKeyPath)

	if err := machineUtil.GenerateCert([]string{""}, serverCertPath, serverKeyPath, caCertPath, caKeyPath, name, bits); err != nil {
		fmt.Println(err.Error())
		return
	}
}
