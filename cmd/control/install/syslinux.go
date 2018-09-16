package install

import (
	"bufio"
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/os/pkg/log"
)

func syslinuxConfig(menu BootVars) error {
	log.Debugf("syslinuxConfig")

	filetmpl, err := template.New("syslinuxconfig").Parse(`{{define "syslinuxmenu"}}
LABEL {{.Name}}
    LINUX ../vmlinuz-{{.Version}}-rancheros
    APPEND {{.KernelArgs}} {{.Append}}
    INITRD ../initrd-{{.Version}}-rancheros
{{end}}
TIMEOUT 20   #2 seconds
DEFAULT RancherOS-current

{{- range .Entries}}
{{template "syslinuxmenu" .}}
{{- end}}

`)
	if err != nil {
		log.Errorf("syslinuxconfig %s", err)
		return err
	}

	cfgFile := filepath.Join(menu.BaseName, menu.BootDir+"syslinux/syslinux.cfg")
	log.Debugf("syslinuxConfig written to %s", cfgFile)
	f, err := os.Create(cfgFile)
	if err != nil {
		log.Errorf("Create(%s) %s", cfgFile, err)
		return err
	}
	err = filetmpl.Execute(f, menu)
	if err != nil {
		return err
	}
	return nil
}

func ReadGlobalCfg(globalCfg string) (string, error) {
	append := ""
	buf, err := ioutil.ReadFile(globalCfg)
	if err != nil {
		return append, err
	}

	s := bufio.NewScanner(bytes.NewReader(buf))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "APPEND") {
			append = strings.TrimSpace(strings.TrimPrefix(line, "APPEND"))
		}
	}
	return append, nil
}

func ReadSyslinuxCfg(currentCfg string) (string, string, error) {
	vmlinuzFile := ""
	initrdFile := ""
	// Need to parse currentCfg for the lines:
	// KERNEL ../vmlinuz-4.9.18-rancher^M
	// INITRD ../initrd-41e02e6-dirty^M
	buf, err := ioutil.ReadFile(currentCfg)
	if err != nil {
		return vmlinuzFile, initrdFile, err
	}

	DIST := filepath.Dir(currentCfg)

	s := bufio.NewScanner(bytes.NewReader(buf))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "KERNEL") {
			vmlinuzFile = strings.TrimSpace(strings.TrimPrefix(line, "KERNEL"))
			vmlinuzFile = filepath.Join(DIST, filepath.Base(vmlinuzFile))
		}
		if strings.HasPrefix(line, "INITRD") {
			initrdFile = strings.TrimSpace(strings.TrimPrefix(line, "INITRD"))
			initrdFile = filepath.Join(DIST, filepath.Base(initrdFile))
		}
	}
	return vmlinuzFile, initrdFile, err
}
