package console

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/cmd/cloudinitexecute"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

const (
	consoleDone = "/run/console-done"
	dockerHome  = "/home/docker"
	rancherHome = "/home/rancher"
	startScript = "/opt/rancher/bin/start.sh"
)

type symlink struct {
	oldname, newname string
}

func Main() {
	cfg := config.LoadConfig()

	if err := os.MkdirAll(rancherHome, 2755); err != nil {
		log.Error(err)
	}
	if err := os.Chown(rancherHome, 1100, 1100); err != nil {
		log.Error(err)
	}

	if err := os.MkdirAll(dockerHome, 2755); err != nil {
		log.Error(err)
	}
	if err := os.Chown(dockerHome, 1101, 1101); err != nil {
		log.Error(err)
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

	respawn := `
/sbin/getty 115200 tty6
/sbin/getty 115200 tty5
/sbin/getty 115200 tty4
/sbin/getty 115200 tty3
/sbin/getty 115200 tty2
/sbin/getty 115200 tty1
/usr/sbin/sshd -D`

	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.Error(err)
	}
	cmdlineString := string(cmdline)

	for _, tty := range []string{"ttyS0", "ttyS1", "ttyS2", "ttyS3", "ttyAMA0"} {
		if strings.Contains(cmdlineString, "console="+tty) {
			respawn += "\n/sbin/getty 115200 " + tty
		}
	}

	if err = ioutil.WriteFile("/etc/respawn.conf", []byte(respawn), 0644); err != nil {
		log.Error(err)
	}

	if err = modifySshdConfig(); err != nil {
		log.Error(err)
	}

	if err = writeOsRelease(); err != nil {
		log.Error(err)
	}

	for _, link := range []symlink{
		{"/var/lib/rancher/engine/docker", "/usr/bin/docker"},
		{"/var/lib/rancher/engine/docker-containerd", "/usr/bin/docker-containerd"},
		{"/var/lib/rancher/engine/docker-containerd-ctr", "/usr/bin/docker-containerd-ctr"},
		{"/var/lib/rancher/engine/docker-containerd-shim", "/usr/bin/docker-containerd-shim"},
		{"/var/lib/rancher/engine/dockerd", "/usr/bin/dockerd"},
		{"/var/lib/rancher/engine/docker-runc", "/usr/bin/docker-runc"},
	} {
		syscall.Unlink(link.newname)
		if err = os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}

	cmd = exec.Command("bash", "-c", `echo 'RancherOS \n \l' > /etc/issue`)
	if err = cmd.Run(); err != nil {
		log.Error(err)
	}

	cmd = exec.Command("bash", "-c", `echo $(/sbin/ifconfig | grep -B1 "inet addr" |awk '{ if ( $1 == "inet" ) { print $2 } else if ( $2 == "Link" ) { printf "%s:" ,$1 } }' |awk -F: '{ print $1 ": " $3}') >> /etc/issue`)
	if err = cmd.Run(); err != nil {
		log.Error(err)
	}

	cloudinitexecute.ApplyConsole(cfg)

	if util.ExistsAndExecutable(config.CloudConfigScriptFile) {
		cmd := exec.Command(config.CloudConfigScriptFile)
		if err = cmd.Run(); err != nil {
			log.Error(err)
		}
	}

	if util.ExistsAndExecutable(startScript) {
		cmd := exec.Command(startScript)
		if err = cmd.Run(); err != nil {
			log.Error(err)
		}
	}

	if util.ExistsAndExecutable("/etc/rc.local") {
		cmd := exec.Command("/etc/rc.local")
		if err = cmd.Run(); err != nil {
			log.Error(err)
		}
	}

	os.Setenv("TERM", "linux")

	respawnBinPath, err := exec.LookPath("respawn")
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(consoleDone, []byte(cfg.Rancher.Console), 0644); err != nil {
		log.Error(err)
	}

	log.Fatal(syscall.Exec(respawnBinPath, []string{"respawn", "-f", "/etc/respawn.conf"}, os.Environ()))
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

	return ioutil.WriteFile("/etc/ssh/ssh_config", []byte(sshdConfigString), 0644)
}

func writeOsRelease() error {
	idLike := "busybox"
	if osRelease, err := ioutil.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(osRelease), "\n") {
			if strings.HasPrefix(line, "ID_LIKE") {
				split := strings.Split(line, "ID_LIKE")
				if len(split) > 1 {
					idLike = split[1]
				}
			}
		}
	}

	return ioutil.WriteFile("/etc/os-release", []byte(fmt.Sprintf(`
NAME="RancherOS"
VERSION=%s
ID=rancheros
ID_LIKE=%s
VERSION_ID=%s
PRETTY_NAME="RancherOS %s"
HOME_URL=
SUPPORT_URL=
BUG_REPORT_URL=
BUILD_ID=
`, config.VERSION, idLike, config.VERSION, config.VERSION)), 0644)
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
