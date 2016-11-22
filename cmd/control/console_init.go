package control

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/os/cmd/cloudinitexecute"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

const (
	consoleDone = "/run/console-done"
	dockerHome  = "/home/docker"
	gettyCmd    = "/sbin/agetty"
	rancherHome = "/home/rancher"
	startScript = "/opt/rancher/bin/start.sh"
)

type symlink struct {
	oldname, newname string
}

func consoleInitAction(c *cli.Context) error {
	cfg := config.LoadConfig()

	if _, err := os.Stat(rancherHome); os.IsNotExist(err) {
		if err := os.MkdirAll(rancherHome, 0755); err != nil {
			log.Error(err)
		}
		if err := os.Chown(rancherHome, 1100, 1100); err != nil {
			log.Error(err)
		}
	}

	if _, err := os.Stat(dockerHome); os.IsNotExist(err) {
		if err := os.MkdirAll(dockerHome, 0755); err != nil {
			log.Error(err)
		}
		if err := os.Chown(dockerHome, 1101, 1101); err != nil {
			log.Error(err)
		}
	}

	password := config.GetCmdline("rancher.password")
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprint("rancher:", password))
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	cmd = exec.Command("bash", "-c", `sed -E -i 's/(rancher:.*:).*(:.*:.*:.*:.*:.*:.*)$/\1\2/' /etc/shadow`)
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	if err := setupSSH(cfg); err != nil {
		log.Error(err)
	}

	if err := writeRespawn(); err != nil {
		log.Error(err)
	}

	if err := modifySshdConfig(); err != nil {
		log.Error(err)
	}

	for _, link := range []symlink{
		{"/var/lib/rancher/engine/docker", "/usr/bin/docker"},
		{"/var/lib/rancher/engine/docker-containerd", "/usr/bin/docker-containerd"},
		{"/var/lib/rancher/engine/docker-containerd-ctr", "/usr/bin/docker-containerd-ctr"},
		{"/var/lib/rancher/engine/docker-containerd-shim", "/usr/bin/docker-containerd-shim"},
		{"/var/lib/rancher/engine/dockerd", "/usr/bin/dockerd"},
		{"/var/lib/rancher/engine/docker-proxy", "/usr/bin/docker-proxy"},
		{"/var/lib/rancher/engine/docker-runc", "/usr/bin/docker-runc"},
		{"/usr/share/rancher/os-release", "/usr/lib/os-release"},
		{"/usr/share/rancher/os-release", "/etc/os-release"},
	} {
		syscall.Unlink(link.newname)
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}

	cmd = exec.Command("bash", "-c", `echo 'RancherOS \n \l' > /etc/issue`)
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	cmd = exec.Command("bash", "-c", `echo $(/sbin/ifconfig | grep -B1 "inet addr" |awk '{ if ( $1 == "inet" ) { print $2 } else if ( $2 == "Link" ) { printf "%s:" ,$1 } }' |awk -F: '{ print $1 ": " $3}') >> /etc/issue`)
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	cloudinitexecute.ApplyConsole(cfg)

	if err := util.RunScript(config.CloudConfigScriptFile); err != nil {
		log.Error(err)
	}
	if err := util.RunScript(startScript); err != nil {
		log.Error(err)
	}

	if err := ioutil.WriteFile(consoleDone, []byte(cfg.Rancher.Console), 0644); err != nil {
		log.Error(err)
	}

	if err := util.RunScript("/etc/rc.local"); err != nil {
		log.Error(err)
	}

	os.Setenv("TERM", "linux")

	respawnBinPath, err := exec.LookPath("respawn")
	if err != nil {
		return err
	}

	return syscall.Exec(respawnBinPath, []string{"respawn", "-f", "/etc/respawn.conf"}, os.Environ())
}

func generateRespawnConf(cmdline string) string {
	var respawnConf bytes.Buffer

	for i := 1; i < 7; i++ {
		tty := fmt.Sprintf("tty%d", i)

		respawnConf.WriteString(gettyCmd)
		if strings.Contains(cmdline, fmt.Sprintf("rancher.autologin=%s", tty)) {
			respawnConf.WriteString(" --autologin rancher")
		}
		respawnConf.WriteString(fmt.Sprintf(" 115200 %s\n", tty))
	}

	for _, tty := range []string{"ttyS0", "ttyS1", "ttyS2", "ttyS3", "ttyAMA0"} {
		if !strings.Contains(cmdline, fmt.Sprintf("console=%s", tty)) {
			continue
		}

		respawnConf.WriteString(gettyCmd)
		if strings.Contains(cmdline, fmt.Sprintf("rancher.autologin=%s", tty)) {
			respawnConf.WriteString(" --autologin rancher")
		}
		respawnConf.WriteString(fmt.Sprintf(" 115200 %s\n", tty))
	}

	respawnConf.WriteString("/usr/sbin/sshd -D")

	return respawnConf.String()
}

func writeRespawn() error {
	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}

	respawn := generateRespawnConf(string(cmdline))

	files, err := ioutil.ReadDir("/etc/respawn.conf.d")
	if err == nil {
		for _, f := range files {
			p := path.Join("/etc/respawn.conf.d", f.Name())
			content, err := ioutil.ReadFile(p)
			if err != nil {
				log.Errorf("Failed to read %s: %v", p, err)
				continue
			}
			respawn += fmt.Sprintf("\n%s", string(content))
		}
	} else if !os.IsNotExist(err) {
		log.Error(err)
	}

	return ioutil.WriteFile("/etc/respawn.conf", []byte(respawn), 0644)
}

func modifySshdConfig() error {
	sshdConfig, err := ioutil.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return err
	}
	sshdConfigString := string(sshdConfig)

	for _, item := range []string{
		"UseDNS no",
		"PermitRootLogin no",
		"ServerKeyBits 2048",
		"AllowGroups docker",
	} {
		match, err := regexp.Match("^"+item, sshdConfig)
		if err != nil {
			return err
		}
		if !match {
			sshdConfigString += fmt.Sprintf("%s\n", item)
		}
	}

	return ioutil.WriteFile("/etc/ssh/sshd_config", []byte(sshdConfigString), 0644)
}

func setupSSH(cfg *config.CloudConfig) error {
	for _, keyType := range []string{"rsa", "dsa", "ecdsa", "ed25519"} {
		outputFile := fmt.Sprintf("/etc/ssh/ssh_host_%s_key", keyType)
		outputFilePub := fmt.Sprintf("/etc/ssh/ssh_host_%s_key.pub", keyType)

		if _, err := os.Stat(outputFile); err == nil {
			continue
		}

		saved, savedExists := cfg.Rancher.Ssh.Keys[keyType]
		pub, pubExists := cfg.Rancher.Ssh.Keys[keyType+"-pub"]

		if savedExists && pubExists {
			// TODO check permissions
			if err := util.WriteFileAtomic(outputFile, []byte(saved), 0600); err != nil {
				return err
			}
			if err := util.WriteFileAtomic(outputFilePub, []byte(pub), 0600); err != nil {
				return err
			}
			continue
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf("ssh-keygen -f %s -N '' -t %s", outputFile, keyType))
		if err := cmd.Run(); err != nil {
			return err
		}

		savedBytes, err := ioutil.ReadFile(outputFile)
		if err != nil {
			return err
		}

		pubBytes, err := ioutil.ReadFile(outputFilePub)
		if err != nil {
			return err
		}

		config.Set(fmt.Sprintf("rancher.ssh.keys.%s", keyType), string(savedBytes))
		config.Set(fmt.Sprintf("rancher.ssh.keys.%s-pub", keyType), string(pubBytes))
	}

	return os.MkdirAll("/var/run/sshd", 0644)
}
