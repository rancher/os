package console

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

func Main() {
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

	if err := writeOsRelease(); err != nil {
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
	} {
		syscall.Unlink(link.newname)
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}

	if err := ioutil.WriteFile("/etc/issue", []byte("RancherOS \\n \\l\n"), 0644); err != nil {
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
		log.Fatal(err)
	}

	log.Fatal(syscall.Exec(respawnBinPath, []string{"respawn", "-f", "/etc/respawn.conf"}, os.Environ()))
}

func generateRespawnConf(cmdline string) string {
	var respawnConf bytes.Buffer

	autologins := make(map[string]struct{})
	for _, v := range config.GetCmdLineValues(cmdline, "rancher.autologin") {
		autologins[v] = struct{}{}
	}

	ttys := append(config.GetCmdLineValues(cmdline, "console"),
		[]string{"tty1", "tty2", "tty3", "tty4", "tty5", "tty6", "tty7"}...)
	unique_ttys := make(map[string]struct{})
	for _, tty := range ttys {
		baudrate := "115200"
		kv := strings.SplitN(tty, ",", 2)
		if len(kv) == 2 {
			tty = kv[0]
			baudrate = kv[1]
		}

		if _, ok := unique_ttys[tty]; ok {
			continue
		}
		_, autologin := autologins[tty]
		respawnConf.WriteString(writeAgetty(cmdline, tty, baudrate, autologin))
		unique_ttys[tty] = struct{}{}
	}

	respawnConf.WriteString("/usr/sbin/sshd -D")
	return respawnConf.String()
}

func writeAgetty(cmdline, tty, baudrate string, autologin bool) string {
	var agettyLine bytes.Buffer

	agettyLine.WriteString(gettyCmd)
	if autologin {
		agettyLine.WriteString(" --autologin rancher")
	}
	agettyLine.WriteString(fmt.Sprintf(" %s %s\n", baudrate, tty))
	return agettyLine.String()
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
