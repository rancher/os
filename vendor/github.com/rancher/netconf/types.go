package netconf

type NetworkConfig struct {
	PreCmds    []string                   `yaml:"pre_cmds,omitempty"`
	Dns        DnsConfig                  `yaml:"dns,omitempty"`
	Interfaces map[string]InterfaceConfig `yaml:"interfaces,omitempty"`
	PostCmds   []string                   `yaml:"post_cmds,omitempty"`
	HttpProxy  string                     `yaml:"http_proxy,omitempty"`
	HttpsProxy string                     `yaml:"https_proxy,omitempty"`
	NoProxy    string                     `yaml:"no_proxy,omitempty"`
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
}

type DnsConfig struct {
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
}
