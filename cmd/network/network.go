package network

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"text/template"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/docker"
	"github.com/burmilla/os/pkg/hostname"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/netconf"

	"github.com/docker/libnetwork/resolvconf"
	"golang.org/x/net/context"
)

var funcMap = template.FuncMap{
	"addFunc": func(a, b int) string {
		return strconv.Itoa(a + b)
	},
}

func Main() {
	log.InitLogger()

	cfg := config.LoadConfig()
	ApplyNetworkConfig(cfg)

	log.Infof("Restart syslog")
	client, err := docker.NewSystemClient()
	if err != nil {
		log.Error(err)
	}

	if err := client.ContainerRestart(context.Background(), "syslog", 10); err != nil {
		log.Error(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	log.Info("Received SIGTERM, shutting down")
	netconf.StopWpaSupplicant()
	netconf.StopDhcpcd()
}

func ApplyNetworkConfig(cfg *config.CloudConfig) {
	log.Infof("Apply Network Config")
	userSetDNS := len(cfg.Rancher.Network.DNS.Nameservers) > 0 || len(cfg.Rancher.Network.DNS.Search) > 0

	if err := hostname.SetHostnameFromCloudConfig(cfg); err != nil {
		log.Errorf("Failed to set hostname from cloud config: %v", err)
	}

	userSetHostname := cfg.Hostname != ""
	if cfg.Rancher.Network.DHCPTimeout <= 0 {
		cfg.Rancher.Network.DHCPTimeout = cfg.Rancher.Defaults.Network.DHCPTimeout
	}

	// In order to handle the STATIC mode in Wi-Fi network, we have to update the dhcpcd.conf file.
	// https://wiki.archlinux.org/index.php/dhcpcd#Static_profile
	if len(cfg.Rancher.Network.WifiNetworks) > 0 {
		generateDhcpcdFiles(cfg)
		generateWpaFiles(cfg)
	}

	dhcpSetDNS, err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network, userSetHostname, userSetDNS)
	if err != nil {
		log.Errorf("Failed to apply network configs(by netconf): %v", err)
	}

	if dhcpSetDNS {
		log.Infof("DNS set by DHCP")
	}

	if !userSetDNS && !dhcpSetDNS {
		// only write 8.8.8.8,8.8.4.4 as a last resort
		log.Infof("Writing default resolv.conf - no user setting, and no DHCP setting")
		if _, err := resolvconf.Build("/etc/resolv.conf",
			cfg.Rancher.Defaults.Network.DNS.Nameservers,
			cfg.Rancher.Defaults.Network.DNS.Search,
			nil); err != nil {
			log.Errorf("Failed to write resolv.conf (!userSetDNS and !dhcpSetDNS): %v", err)
		}
	}
	if userSetDNS {
		if _, err := resolvconf.Build("/etc/resolv.conf", cfg.Rancher.Network.DNS.Nameservers, cfg.Rancher.Network.DNS.Search, nil); err != nil {
			log.Errorf("Failed to write resolv.conf (userSetDNS): %v", err)
		} else {
			log.Infof("writing to /etc/resolv.conf: nameservers: %v, search: %v", cfg.Rancher.Network.DNS.Nameservers, cfg.Rancher.Network.DNS.Search)
		}
	}

	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Debugf("Resolve.conf == [%s], %v", resolve, err)

	log.Infof("Apply Network Config SyncHostname")
	if err := hostname.SyncHostname(); err != nil {
		log.Errorf("Failed to sync hostname: %v", err)
	}
}

func generateDhcpcdFiles(cfg *config.CloudConfig) {
	networks := cfg.Rancher.Network.WifiNetworks
	interfaces := cfg.Rancher.Network.Interfaces
	configs := make(map[string]netconf.WifiNetworkConfig)
	for k, v := range interfaces {
		if c, ok := networks[v.WifiNetwork]; ok && c.Address != "" {
			configs[k] = c
		}
	}
	f, err := os.Create(config.DHCPCDConfigFile)
	defer f.Close()
	if err != nil {
		log.Errorf("Failed to open file: %s err: %v", config.DHCPCDConfigFile, err)
	}
	templateFiles := []string{config.DHCPCDTemplateFile}
	templateName := filepath.Base(templateFiles[0])
	p := template.Must(template.New(templateName).ParseFiles(templateFiles...))
	if err = p.Execute(f, configs); err != nil {
		log.Errorf("Failed to wrote wpa configuration to %s: %v", config.DHCPCDConfigFile, err)
	}
}

func generateWpaFiles(cfg *config.CloudConfig) {
	networks := cfg.Rancher.Network.WifiNetworks
	interfaces := cfg.Rancher.Network.Interfaces
	for k, v := range interfaces {
		if v.WifiNetwork != "" {
			configs := make(map[string]netconf.WifiNetworkConfig)
			filename := fmt.Sprintf(config.WPAConfigFile, k)
			f, err := os.Create(filename)
			if err != nil {
				log.Errorf("Failed to open file: %s err: %v", filename, err)
			}
			if c, ok := networks[v.WifiNetwork]; ok {
				configs[v.WifiNetwork] = c
			}
			templateFiles := []string{config.WPATemplateFile}
			templateName := filepath.Base(templateFiles[0])
			p := template.Must(template.New(templateName).Funcs(funcMap).ParseFiles(templateFiles...))
			if err = p.Execute(f, configs); err != nil {
				log.Errorf("Failed to wrote wpa configuration to %s: %v", filename, err)
			}
			f.Close()
		}
	}
}
