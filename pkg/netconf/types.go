package netconf

type NetworkConfig struct {
	PreCmds       []string                      `yaml:"pre_cmds,omitempty"`
	DHCPTimeout   int                           `yaml:"dhcp_timeout,omitempty"`
	DNS           DNSConfig                     `yaml:"dns,omitempty"`
	Interfaces    map[string]InterfaceConfig    `yaml:"interfaces,omitempty"`
	PostCmds      []string                      `yaml:"post_cmds,omitempty"`
	HTTPProxy     string                        `yaml:"http_proxy,omitempty"`
	HTTPSProxy    string                        `yaml:"https_proxy,omitempty"`
	NoProxy       string                        `yaml:"no_proxy,omitempty"`
	WifiNetworks  map[string]WifiNetworkConfig  `yaml:"wifi_networks,omitempty"`
	ModemNetworks map[string]ModemNetworkConfig `yaml:"modem_networks,omitempty"`
}

type InterfaceConfig struct {
	Match       string            `yaml:"match,omitempty"`
	DHCP        bool              `yaml:"dhcp,omitempty"`
	DHCPArgs    string            `yaml:"dhcp_args,omitempty"`
	Address     string            `yaml:"address,omitempty"`
	Addresses   []string          `yaml:"addresses,omitempty"`
	IPV4LL      bool              `yaml:"ipv4ll,omitempty"`
	Gateway     string            `yaml:"gateway,omitempty"`
	GatewayIpv6 string            `yaml:"gateway_ipv6,omitempty"`
	MTU         int               `yaml:"mtu,omitempty"`
	Bridge      string            `yaml:"bridge,omitempty"`
	Bond        string            `yaml:"bond,omitempty"`
	BondOpts    map[string]string `yaml:"bond_opts,omitempty"`
	PostUp      []string          `yaml:"post_up,omitempty"`
	PreUp       []string          `yaml:"pre_up,omitempty"`
	Vlans       string            `yaml:"vlans,omitempty"`
	WifiNetwork string            `yaml:"wifi_network,omitempty"`
}

type DNSConfig struct {
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
}

type WifiNetworkConfig struct {
	Address           string   `yaml:"address,omitempty"`
	Gateway           string   `yaml:"gateway,omitempty"`
	ScanSSID          int      `yaml:"scan_ssid,omitempty"`
	SSID              string   `yaml:"ssid,omitempty"`
	PSK               string   `yaml:"psk,omitempty"`
	Priority          int      `yaml:"priority,omitempty"`
	Pairwise          string   `yaml:"pairwise,omitempty"`
	Group             string   `yaml:"group,omitempty"`
	Eap               string   `yaml:"eap,omitempty"`
	Identity          string   `yaml:"identity,omitempty"`
	AnonymousIdentity string   `yaml:"anonymous_identity,omitempty"`
	CaCerts           []string `yaml:"ca_certs,omitempty"`
	ClientCerts       []string `yaml:"client_certs,omitempty"`
	PrivateKeys       []string `yaml:"private_keys,omitempty"`
	PrivateKeyPasswds []string `yaml:"private_key_passwds,omitempty"`
	Phases            []string `yaml:"phases,omitempty"`
	EapolFlags        int      `yaml:"eapol_flags,omitempty"`
	KeyMgmt           string   `yaml:"key_mgmt,omitempty"`
	Password          string   `yaml:"password,omitempty"`
}

type ModemNetworkConfig struct {
	Apn       string `yaml:"apn"`
	ExtraArgs string `yaml:"extra_args,omitempty"`
}
