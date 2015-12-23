package netconf

type NetworkConfig struct {
	Script     string                     `yaml:"script,omitempty"`
	Dns        DnsConfig                  `yaml:"dns,omitempty"`
	Interfaces map[string]InterfaceConfig `yaml:"interfaces,omitempty"`
}

type InterfaceConfig struct {
	Match       string            `yaml:"match,omitempty"`
	DHCP        bool              `yaml:"dhcp,omitempty"`
	Address     string            `yaml:"address,omitempty"`
	Addresses   []string          `yaml:"addresses,omitempty"`
	IPV4LL      bool              `yaml:"ipv4ll,omitempty"`
	Gateway     string            `yaml:"gateway,omitempty"`
	GatewayIpv6 string            `yaml:"gateway_ipv6,omitempty"`
	MTU         int               `yaml:"mtu,omitempty"`
	Bridge      bool              `yaml:"bridge,omitempty"`
	Bond        string            `yaml:"bond,omitempty"`
	BondOpts    map[string]string `yaml:"bond_opts,omitempty"`
	PostUp      []string          `yaml:"post_up,omitempty"`
}

type DnsConfig struct {
	Override    bool     `yaml:"override"`
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
}
