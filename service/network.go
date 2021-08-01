package service

import (
	"encoding/json"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/netplan"
	"github.com/projectxpolaris/youplus/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net"
	"os/exec"
)

var DefaultNetworkManager = &NetworkManager{}
var IgnoreNetworkName = []string{"lo"}

type IPv4Config struct {
	DHCP    bool     `json:"dhcp"`
	Address []string `json:"address"`
}
type IPv6Config struct {
	DHCP    bool     `json:"dhcp"`
	Address []string `json:"address"`
}
type NetworkInterface struct {
	Name         string              `json:"name"`
	IPv4Address  []string            `json:"IPv4Address"`
	IPv6Address  []string            `json:"IPv6Address"`
	IPv4         IPv4Config          `json:"IPv4"`
	IPv6         IPv6Config          `json:"IPv6"`
	HardwareInfo NetworkHardwareInfo `json:"hardwareInfo"`
}
type NetworkHardwareInfo struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Claimed       bool   `json:"claimed"`
	Handle        string `json:"handle"`
	Description   string `json:"description"`
	Product       string `json:"product"`
	Vendor        string `json:"vendor"`
	Physid        string `json:"physid"`
	Businfo       string `json:"businfo"`
	Logicalname   string `json:"logicalname"`
	Version       string `json:"version"`
	Serial        string `json:"serial"`
	Units         string `json:"units"`
	Size          int    `json:"size"`
	Capacity      int    `json:"capacity"`
	Width         int    `json:"width"`
	Clock         int    `json:"clock"`
	Configuration struct {
		Autonegotiation string `json:"autonegotiation"`
		Broadcast       string `json:"broadcast"`
		Driver          string `json:"driver"`
		Driverversion   string `json:"driverversion"`
		Duplex          string `json:"duplex"`
		Firmware        string `json:"firmware"`
		IP              string `json:"ip"`
		Latency         string `json:"latency"`
		Link            string `json:"link"`
		Multicast       string `json:"multicast"`
		Port            string `json:"port"`
		Speed           string `json:"speed"`
	} `json:"configuration"`
	Capabilities struct {
		Pm              string `json:"pm"`
		Msi             string `json:"msi"`
		Pciexpress      string `json:"pciexpress"`
		Msix            string `json:"msix"`
		BusMaster       string `json:"bus_master"`
		CapList         string `json:"cap_list"`
		Ethernet        bool   `json:"ethernet"`
		Physical        string `json:"physical"`
		Tp              string `json:"tp"`
		Mii             string `json:"mii"`
		TenBt           string `json:"10bt"`
		TenBtFd         string `json:"10bt-fd"`
		HundredBt       string `json:"100bt"`
		HundredBtFd     string `json:"100bt-fd"`
		ThousandBtFd    string `json:"1000bt-fd"`
		Autonegotiation string `json:"autonegotiation"`
	} `json:"capabilities"`
}
type NetworkManager struct {
	Interfaces []*NetworkInterface
}

func (m *NetworkManager) Init() error {
	if utils.IsFileExist(config.Config.NetConfig) {
		return nil
	}
	netplanConf := netplan.NetPlanConf{Network: netplan.Network{
		Renderer:  "NetworkManager",
		Version:   "2",
		Ethernets: map[string]netplan.Ethernet{},
	}}
	for _, iface := range m.Interfaces {
		netplanConf.Network.Ethernets[iface.Name] = netplan.Ethernet{
			DHCP4:       true,
			DHCP6:       true,
			Optional:    true,
			Nameservers: netplan.Nameserver{},
		}
	}
	err := utils.WriteYaml(netplanConf, config.Config.NetConfig)
	if err != nil {
		return err
	}
	return nil
}
func getNetworkInterfaceList() ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := make([]net.Interface, 0)
	for _, iface := range ifaces {
		isIgnore := false
		for _, ignore := range IgnoreNetworkName {
			if ignore == iface.Name {
				isIgnore = true
				break
			}
		}
		if !isIgnore {
			result = append(result, iface)
		}

	}
	return result, nil
}

func (m *NetworkManager) GetHardwareInfo() ([]NetworkHardwareInfo, error) {
	cmd := exec.Command("lshw", "-C", "network", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var hardwareInfos []NetworkHardwareInfo
	err = json.Unmarshal(output, &hardwareInfos)
	if err != nil {
		return nil, err
	}
	return hardwareInfos, nil
}
func (m *NetworkManager) Load() error {
	m.Interfaces = []*NetworkInterface{}
	// load hardware info
	hardwareInfos, err := m.GetHardwareInfo()
	if err != nil {
		return err
	}
	for _, hardwareInfo := range hardwareInfos {
		m.Interfaces = append(m.Interfaces, &NetworkInterface{
			Name:         hardwareInfo.Logicalname,
			HardwareInfo: hardwareInfo,
			IPv4Address:  []string{},
			IPv6Address:  []string{},
		})
	}
	err = m.Init()
	if err != nil {
		return err
	}
	// load netplan config file
	netConf := netplan.NetPlanConf{}
	yamlFile, err := ioutil.ReadFile(config.Config.NetConfig)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &netConf)
	if err != nil {
		return err
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, networkInterface := range m.Interfaces {
		// read config
		conf, exist := netConf.Network.Ethernets[networkInterface.Name]
		if !exist {
			continue
		}
		networkInterface.IPv4 = IPv4Config{
			DHCP:    conf.DHCP4,
			Address: conf.Addresses,
		}
		networkInterface.IPv6 = IPv6Config{
			DHCP:    conf.DHCP6,
			Address: conf.Addresses,
		}
		// find address
		var iface *net.Interface
		for _, findInterface := range ifaces {
			if findInterface.Name == networkInterface.Name {
				iface = &findInterface
			}
		}
		if iface != nil {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					continue
				}
				if ip.To4() != nil {
					networkInterface.IPv4Address = append(networkInterface.IPv4Address, addr.String())
					continue
				}
				if ip.To16() != nil {
					networkInterface.IPv6Address = append(networkInterface.IPv6Address, addr.String())
				}
			}
		}
	}
	return nil
}
