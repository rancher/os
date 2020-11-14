package hypervisor

import (
	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"
)

func Tools(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	enableHypervisorService(cfg, util.GetHypervisor())
	return config.LoadConfig(), nil
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
