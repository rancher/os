package util

import (
	"archive/tar"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/mount"
	machine_utils "github.com/docker/machine/utils"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func TLSConf() {
	name := "rancher"
	bits := 2048

	vargs := os.Args

	caCertPath := "ca.pem"
	caKeyPath := "ca-key.pem"
	outDir := "/var/run/"
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
		if err := machine_utils.GenerateCACertificate(caCertPath, caKeyPath, name, bits); err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		if inputCaKey == "" || inputCaCert == "" {
			fmt.Println("Please specify caKey and CaCert along with -g")
			return
		}

		if _, err := os.Stat(inputCaKey); err != nil {
			//throw error if input ca key not found
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

	if err := machine_utils.GenerateCert([]string{""}, serverCertPath, serverKeyPath, caCertPath, caKeyPath, name, bits); err != nil {
		fmt.Println(err.Error())
		return
	}

}

func mountProc() error {
	if _, err := os.Stat("/proc/self/mountinfo"); os.IsNotExist(err) {
		if _, err := os.Stat("/proc"); os.IsNotExist(err) {
			if err = os.Mkdir("/proc", 0755); err != nil {
				return err
			}
		}

		if err := syscall.Mount("none", "/proc", "proc", 0, ""); err != nil {
			return err
		}
	}

	return nil
}

func Mount(device, directory, fsType, options string) error {
	if err := mountProc(); err != nil {
		return nil
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	return mount.Mount(device, directory, fsType, options)
}

func Remount(directory, options string) error {
	return mount.Mount("", directory, "", fmt.Sprintf("remount,%s", options))
}

func ExtractTar(archive string, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	input := tar.NewReader(f)

	for {
		header, err := input.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if header == nil {
			break
		}

		fileInfo := header.FileInfo()
		fileName := path.Join(dest, header.Name)
		if fileInfo.IsDir() {
			//log.Debugf("DIR : %s", fileName)
			err = os.MkdirAll(fileName, fileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			//log.Debugf("FILE: %s", fileName)
			destFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(destFile, input)
			// Not deferring, concerned about holding open too many files
			destFile.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Contains(values []string, value string) bool {
	if len(value) == 0 {
		return false
	}

	for _, i := range values {
		if i == value {
			return true
		}
	}

	return false
}

type ReturnsErr func() error

func ShortCircuit(funcs ...ReturnsErr) error {
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}

type ErrWriter struct {
	w   io.Writer
	Err error
}

func NewErrorWriter(w io.Writer) *ErrWriter {
	return &ErrWriter{
		w: w,
	}
}

func (e *ErrWriter) Write(buf []byte) *ErrWriter {
	if e.Err != nil {
		return e
	}

	_, e.Err = e.w.Write(buf)
	return e
}

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
