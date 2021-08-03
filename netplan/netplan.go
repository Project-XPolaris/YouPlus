package netplan

type Network struct {
	Renderer  string              `yaml:"renderer,omitempty"`
	Version   string              `yaml:"version,omitempty"`
	Ethernets map[string]Ethernet `yaml:"ethernets,omitempty"`
}
type Ethernet struct {
	Addresses   []string   `yaml:"addresses,omitempty"`
	DHCP4       bool       `yaml:"dhcp4"`
	DHCP6       bool       `yaml:"dhcp6"`
	Optional    bool       `yaml:"optional"`
	Gateway4    string     `yaml:"gateway4,omitempty"`
	Gateway6    string     `yaml:"gateway6,omitempty"`
	Nameservers Nameserver `yaml:"nameservers,omitempty"`
}
type Nameserver struct {
	Addresses []string `yaml:"addresses,omitempty"`
}
type NetPlanConf struct {
	Network Network `yaml:"network,omitempty"`
}
