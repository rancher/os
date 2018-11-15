package hypervisor

import (
	"io/ioutil"
	"os"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
)

func Tools(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	enableHypervisorService(cfg, util.GetHypervisor())
	return config.LoadConfig(), nil
}

func Env(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	hypervisor := util.GetHypervisor()

	// this code make sure the open-vm-tools service can be started correct when there is no network
	if hypervisor == "vmware" {
		// make sure the cache directory exist
		if err := os.MkdirAll("/var/lib/rancher/cache/", os.ModeDir|0755); err != nil {
			log.Errorf("Create service cache diretory error: %v", err)
		}

		// move os-services cache file
		if _, err := os.Stat("/usr/share/ros/services-cache"); err == nil {
			files, err := ioutil.ReadDir("/usr/share/ros/services-cache/")
			if err != nil {
				log.Errorf("Read file error: %v", err)
			}
			for _, f := range files {
				err := os.Rename("/usr/share/ros/services-cache/"+f.Name(), "/var/lib/rancher/cache/"+f.Name())
				if err != nil {
					log.Errorf("Rename file error: %v", err)
				}
			}
			if err := os.Remove("/usr/share/ros/services-cache"); err != nil {
				log.Errorf("Remove file error: %v", err)
			}
		}

	}

	return cfg, nil
}

func enableHypervisorService(cfg *config.CloudConfig, hypervisorName string) {
	if hypervisorName == "" {
		return
	}

	// enable open-vm-tools and hyperv-vm-tools
	// these services(xenhvm-vm-tools, kvm-vm-tools, and bhyve-vm-tools) don't exist yet
	serviceName := ""
	switch hypervisorName {
	case "vmware":
		serviceName = "open-vm-tools"
	case "hyperv":
		serviceName = "hyperv-vm-tools"
	default:
		log.Infof("no hypervisor matched")
	}

	if serviceName != "" {
		if !cfg.Rancher.HypervisorService {
			log.Infof("Skipping %s as `rancher.hypervisor_service` is set to false", serviceName)
			return
		}

		// Check removed - there's an x509 cert failure on first boot of an installed system
		// check quickly to see if there is a yml file available
		//	if service.ValidService(serviceName, cfg) {
		log.Infof("Setting rancher.services_include. %s=true", serviceName)
		if err := config.Set("rancher.services_include."+serviceName, "true"); err != nil {
			log.Error(err)
		}
	}
}
