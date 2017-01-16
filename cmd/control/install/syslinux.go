package install

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/rancher/os/log"
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
