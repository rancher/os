package control

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/burmilla/os/cmd/cloudinitexecute"
	"github.com/burmilla/os/config"
	"github.com/burmilla/os/config/cmdline"
	"github.com/burmilla/os/pkg/compose"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"

	"github.com/codegangsta/cli"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
)

const (
	consoleDone = "/run/console-done"
	dockerHome  = "/home/docker"
	gettyCmd    = "/sbin/agetty"
	rancherHome = "/home/rancher"
	startScript = "/opt/rancher/bin/start.sh"
	runLockDir  = "/run/lock"
	sshdFile    = "/etc/ssh/sshd_config"
	sshdTplFile = "/etc/ssh/sshd_config.tpl"
)

type symlink struct {
	oldname, newname string
}

func consoleInitAction(c *cli.Context) error {
	return consoleInitFunc()
}

func createHomeDir(homedir string, uid, gid int) {
	if _, err := os.Stat(homedir); os.IsNotExist(err) {
		if err := os.MkdirAll(homedir, 0755); err != nil {
			log.Error(err)
		}
		if err := os.Chown(homedir, uid, gid); err != nil {
			log.Error(err)
		}
	}
}

func enableBashRC(homedir string, uid, gid int) {
	if _, err := os.Stat(homedir + "/.bash_logout"); os.IsNotExist(err) {
		if err := util.FileCopy("/etc/skel/.bash_logout", homedir+"/.bash_logout"); err != nil {
			log.Error(err)
		}
		if err := os.Chown(homedir+"/.bash_logout", uid, gid); err != nil {
			log.Error(err)
		}
	}

	if _, err := os.Stat(homedir + "/.bashrc"); os.IsNotExist(err) {
		if err := util.FileCopy("/etc/skel/.bashrc", homedir+"/.bashrc"); err != nil {
			log.Error(err)
		}
		if err := os.Chown(homedir+"/.bashrc", uid, gid); err != nil {
			log.Error(err)
		}
	}

	if _, err := os.Stat(homedir + "/.profile"); os.IsNotExist(err) {
		if err := util.FileCopy("/etc/skel/.profile", homedir+"/.profile"); err != nil {
			log.Error(err)
		}
		if err := os.Chown(homedir+"/.profile", uid, gid); err != nil {
			log.Error(err)
		}
	}
}

func consoleInitFunc() error {
	cfg := config.LoadConfig()

	// Now that we're booted, stop writing debug messages to the console
	cmd := exec.Command("sudo", "dmesg", "--console-off")
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	createHomeDir(rancherHome, 1100, 1100)
	createHomeDir(dockerHome, 1101, 1101)

	// who & w command need this file
	if _, err := os.Stat("/run/utmp"); os.IsNotExist(err) {
		f, err := os.OpenFile("/run/utmp", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Error(err)
		}
		defer f.Close()
	}

	// last command need this file
	if _, err := os.Stat("/var/log/wtmp"); os.IsNotExist(err) {
		f, err := os.OpenFile("/var/log/wtmp", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Error(err)
		}
		defer f.Close()
	}

	// some software need this dir, like open-iscsi
	if _, err := os.Stat(runLockDir); os.IsNotExist(err) {
		if err = os.Mkdir(runLockDir, 0755); err != nil {
			log.Error(err)
		}
	}

	ignorePassword := false
	for _, d := range cfg.Rancher.Disable {
		if d == "password" {
			ignorePassword = true
			break
		}
	}

	password := cmdline.GetCmdline("rancher.password")
	if !ignorePassword && password != "" {
		cmd := exec.Command("chpasswd")
		cmd.Stdin = strings.NewReader(fmt.Sprint("rancher:", password))
		if err := cmd.Run(); err != nil {
			log.Error(err)
		}

		cmd = exec.Command("bash", "-c", `sed -E -i 's/(rancher:.*:).*(:.*:.*:.*:.*:.*:.*)$/\1\2/' /etc/shadow`)
		if err := cmd.Run(); err != nil {
			log.Error(err)
		}
	}

	if err := setupSSH(cfg); err != nil {
		log.Error(err)
	}

	if err := writeRespawn("rancher", cfg.Rancher.SSH.Daemon, false); err != nil {
		log.Error(err)
	}

	if err := modifySshdConfig(cfg); err != nil {
		log.Error(err)
	}

	p, err := compose.GetProject(cfg, false, true)
	if err != nil {
		log.Error(err)
	}

	// check the multi engine service & generate the multi engine script
	for _, key := range p.ServiceConfigs.Keys() {
		serviceConfig, ok := p.ServiceConfigs.Get(key)
		if !ok {
			log.Errorf("Failed to get service config from the project")
			continue
		}
		if _, ok := serviceConfig.Labels[config.UserDockerLabel]; ok {
			err = util.GenerateDindEngineScript(serviceConfig.Labels[config.UserDockerLabel])
			if err != nil {
				log.Errorf("Failed to generate engine script: %v", err)
				continue
			}
		}
	}

	baseSymlink := symLinkEngineBinary()

	if _, err := os.Stat(dockerCompletionFile); err == nil {
		baseSymlink = append(baseSymlink, symlink{
			dockerCompletionFile, dockerCompletionLinkFile,
		})
	}

	// create placeholder for docker-compose binary
	const ComposePlaceholder = `
#!/bin/bash
echo 'INFO: System service "docker-compose" is not yet enabled'
sudo ros service enable docker-compose
sudo ros service up docker-compose
`
	if _, err := os.Stat("/var/lib/rancher/compose"); os.IsNotExist(err) {
		if err := os.MkdirAll("/var/lib/rancher/compose", 0555); err != nil {
			log.Error(err)
		}
	}
	if _, err := os.Stat("/var/lib/rancher/compose/docker-compose"); os.IsNotExist(err) {
		if err := ioutil.WriteFile("/var/lib/rancher/compose/docker-compose", []byte(ComposePlaceholder), 0755); err != nil {
			log.Error(err)
		}
	}

	for _, link := range baseSymlink {
		syscall.Unlink(link.newname)
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}

	// mount systemd cgroups
	if err := os.MkdirAll("/sys/fs/cgroup/systemd", 0555); err != nil {
		log.Error(err)
	}
	if err := unix.Mount("cgroup", "/sys/fs/cgroup/systemd", "cgroup", 0, "none,name=systemd"); err != nil {
		log.Error(err)
	}

	// font backslashes need to be escaped for when issue is output! (but not the others..)
	if err := ioutil.WriteFile("/etc/issue", []byte(config.Banner), 0644); err != nil {
		log.Error(err)
	}

	// write out a profile.d file for the proxy settings.
	// maybe write these on the host and bindmount into everywhere?
	proxyLines := []string{}
	for _, k := range []string{"http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY", "no_proxy", "NO_PROXY"} {
		if v, ok := cfg.Rancher.Environment[k]; ok {
			proxyLines = append(proxyLines, fmt.Sprintf("export %s=%q", k, v))
		}
	}

	if len(proxyLines) > 0 {
		proxyString := strings.Join(proxyLines, "\n")
		proxyString = fmt.Sprintf("#!/bin/sh\n%s\n", proxyString)
		if err := ioutil.WriteFile("/etc/profile.d/proxy.sh", []byte(proxyString), 0755); err != nil {
			log.Error(err)
		}
	}

	// write out a profile.d file for the PATH settings.
	pathLines := []string{}
	for _, k := range []string{"PATH", "path"} {
		if v, ok := cfg.Rancher.Environment[k]; ok {
			for _, p := range strings.Split(v, ",") {
				pathLines = append(pathLines, fmt.Sprintf("export PATH=$PATH:%s", strings.TrimSpace(p)))
			}
		}
	}
	if len(pathLines) > 0 {
		pathString := strings.Join(pathLines, "\n")
		pathString = fmt.Sprintf("#!/bin/sh\n%s\n", pathString)
		if err := ioutil.WriteFile("/etc/profile.d/path.sh", []byte(pathString), 0755); err != nil {
			log.Error(err)
		}
	}

	cmd = exec.Command("bash", "-c", `echo $(/sbin/ifconfig | grep -B1 "inet" |awk '{ if ( $1 == "inet" ) { print $2 } else if ( $3 == "mtu" ) { printf "%s:" ,$1 } }' |awk -F: '{ print $1 ": " $3}') >> /etc/issue`)
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

	if err := ioutil.WriteFile(consoleDone, []byte(CurrentConsole()), 0644); err != nil {
		log.Error(err)
	}

	if err := util.RunScript("/etc/rc.local"); err != nil {
		log.Error(err)
	}

	// Enable Bash colors
	enableBashRC("/root", 0, 0)
	enableBashRC(rancherHome, 1100, 1100)
	enableBashRC(dockerHome, 1101, 1101)

	os.Setenv("TERM", "linux")

	respawnBinPath, err := exec.LookPath("respawn")
	if err != nil {
		return err
	}

	return syscall.Exec(respawnBinPath, []string{"respawn", "-f", "/etc/respawn.conf"}, os.Environ())
}

func generateRespawnConf(cmdline, user string, sshd, recovery bool) string {
	var respawnConf bytes.Buffer

	autologinBin := "/usr/bin/autologin"
	if recovery {
		autologinBin = "/usr/bin/recovery"
	}

	config := config.LoadConfig()

	allowAutoLogin := true
	for _, d := range config.Rancher.Disable {
		if d == "autologin" {
			allowAutoLogin = false
			break
		}
	}

	for i := 1; i < 7; i++ {
		tty := fmt.Sprintf("tty%d", i)
		if !istty(tty) {
			continue
		}

		respawnConf.WriteString(gettyCmd)
		if allowAutoLogin && strings.Contains(cmdline, fmt.Sprintf("rancher.autologin=%s", tty)) {
			respawnConf.WriteString(fmt.Sprintf(" -n -l %s -o %s:tty%d", autologinBin, user, i))
		}
		respawnConf.WriteString(fmt.Sprintf(" --noclear %s linux\n", tty))
	}

	for _, tty := range []string{"ttyS0", "ttyS1", "ttyS2", "ttyS3", "ttyAMA0"} {
		if !strings.Contains(cmdline, fmt.Sprintf("console=%s", tty)) {
			continue
		}

		if !istty(tty) {
			continue
		}

		respawnConf.WriteString(gettyCmd)
		if allowAutoLogin && strings.Contains(cmdline, fmt.Sprintf("rancher.autologin=%s", tty)) {
			respawnConf.WriteString(fmt.Sprintf(" -n -l %s -o %s:%s", autologinBin, user, tty))
		}
		respawnConf.WriteString(fmt.Sprintf(" %s\n", tty))
	}

	if sshd {
		respawnConf.WriteString("/usr/sbin/sshd -D")
	}

	return respawnConf.String()
}

func writeRespawn(user string, sshd, recovery bool) error {
	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}

	respawn := generateRespawnConf(string(cmdline), user, sshd, recovery)

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

func modifySshdConfig(cfg *config.CloudConfig) error {
	_, err := os.Stat(sshdTplFile)
	if err == nil {
		os.Remove(sshdFile)
		sshdTpl, err := template.ParseFiles(sshdTplFile)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(sshdFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		config := map[string]string{}
		if cfg.Rancher.SSH.Port > 0 && cfg.Rancher.SSH.Port < 65355 {
			config["Port"] = strconv.Itoa(cfg.Rancher.SSH.Port)
		}
		if cfg.Rancher.SSH.ListenAddress != "" {
			config["ListenAddress"] = cfg.Rancher.SSH.ListenAddress
		}

		return sshdTpl.Execute(f, config)
	} else if os.IsNotExist(err) {
		return nil
	}

	return err
}

func setupSSH(cfg *config.CloudConfig) error {
	for _, keyType := range []string{"rsa", "dsa", "ecdsa", "ed25519"} {
		outputFile := fmt.Sprintf("/etc/ssh/ssh_host_%s_key", keyType)
		outputFilePub := fmt.Sprintf("/etc/ssh/ssh_host_%s_key.pub", keyType)

		if _, err := os.Stat(outputFile); err == nil {
			continue
		}

		saved, savedExists := cfg.Rancher.SSH.Keys[keyType]
		pub, pubExists := cfg.Rancher.SSH.Keys[keyType+"-pub"]

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

func istty(name string) bool {
	if f, err := os.Open(fmt.Sprintf("/dev/%s", name)); err == nil {
		return terminal.IsTerminal(int(f.Fd()))
	}
	return false
}
