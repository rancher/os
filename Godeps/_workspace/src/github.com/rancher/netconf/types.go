package netconf

type NetworkConfig struct {
	Dns        DnsConfig                  `yaml:"dns,omitempty"`
	Interfaces map[string]InterfaceConfig `yaml:"interfaces,omitempty"`
}

type InterfaceConfig struct {
	Match   string `yaml:"match,omitempty"`
	DHCP    bool   `yaml:"dhcp,omitempty"`
	Address string `yaml:"address,omitempty"`
	IPV4LL  bool   `yaml:"ipv4ll,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
	MTU     int    `yaml:"mtu,omitempty"`
	Bridge  bool   `yaml:"bridge,omitempty"`
}

type DnsConfig struct {
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
}
