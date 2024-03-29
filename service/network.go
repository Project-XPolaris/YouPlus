package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/netplan"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net"
	"os/exec"
)

var DefaultNetworkManager = &NetworkManager{}
var (
	NetworkNotFoundError = errors.New("network interface not found")
)

type IPv4Config struct {
	DHCP    *bool    `json:"dhcp"`
	Address []string `json:"address"`
}
type IPv6Config struct {
	DHCP    *bool    `json:"dhcp"`
	Address []string `json:"address"`
}
type NetworkInterface struct {
	Name         string              `json:"name"`
	IPv4Address  []string            `json:"IPv4Address"`
	IPv6Address  []string            `json:"IPv6Address"`
	IPv4Config   *IPv4Config         `json:"IPv4Config"`
	IPv6Config   *IPv6Config         `json:"IPv6Config"`
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
		Renderer:  utils.GetStringPtr("NetworkManager"),
		Version:   utils.GetStringPtr("2"),
		Ethernets: map[string]netplan.Ethernet{},
	}}
	for _, iface := range m.Interfaces {
		netplanConf.Network.Ethernets[iface.Name] = netplan.Ethernet{
			DHCP4:       utils.GetBoolPtr(true),
			DHCP6:       utils.GetBoolPtr(true),
			Optional:    utils.GetBoolPtr(true),
			Nameservers: netplan.Nameserver{},
		}
	}
	err := utils.WriteYaml(netplanConf, config.Config.NetConfig)
	if err != nil {
		return err
	}
	return nil
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

		networkInterface.IPv4Config = &IPv4Config{
			DHCP:    conf.DHCP4,
			Address: []string{},
		}
		networkInterface.IPv6Config = &IPv6Config{
			DHCP:    conf.DHCP6,
			Address: []string{},
		}
		for _, address := range conf.Addresses {
			ip, _, _ := net.ParseCIDR(address)
			if ip.To4() != nil {
				networkInterface.IPv4Config.Address = append(networkInterface.IPv4Config.Address, address)
			}
			test := ip.To16()
			logrus.Info(test)
			if ip.To16() != nil {
				networkInterface.IPv6Config.Address = append(networkInterface.IPv6Config.Address, address)
			}
		}
		// find address
		var iface *net.Interface
		for _, findInterface := range ifaces {
			if findInterface.Name == networkInterface.Name {
				iface = &findInterface
				break
			}
		}
		if iface != nil {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				fmt.Println(addr.String())
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
func (m *NetworkManager) writeConfig() error {
	netplanConf := netplan.NetPlanConf{Network: netplan.Network{
		Renderer:  utils.GetStringPtr("NetworkManager"),
		Version:   utils.GetStringPtr("2"),
		Ethernets: map[string]netplan.Ethernet{},
	}}
	for _, iface := range m.Interfaces {
		if iface.IPv4Config == nil || iface.IPv6Config == nil {
			continue
		}
		netConf := netplan.Ethernet{
			DHCP4:       iface.IPv4Config.DHCP,
			DHCP6:       iface.IPv6Config.DHCP,
			Optional:    utils.GetBoolPtr(true),
			Nameservers: netplan.Nameserver{},
		}
		configAddrs := make([]string, 0)
		if iface.IPv4Config.Address != nil {
			configAddrs = append(configAddrs, iface.IPv4Config.Address...)
		}
		if iface.IPv6Config.Address != nil {
			configAddrs = append(configAddrs, iface.IPv6Config.Address...)
		}
		if len(configAddrs) > 0 {
			netConf.Addresses = configAddrs
		}
		netplanConf.Network.Ethernets[iface.Name] = netConf
	}
	err := utils.WriteYaml(netplanConf, config.Config.NetConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *NetworkManager) getNetworkByName(name string) *NetworkInterface {
	for _, networkInterface := range m.Interfaces {
		if networkInterface.Name == name {
			return networkInterface
		}
	}
	return nil
}

func (m *NetworkManager) UpdateConfig(name string, ipv4 *IPv4Config, ipv6 *IPv6Config) error {
	network := m.getNetworkByName(name)
	if network == nil {
		return NetworkNotFoundError
	}
	if ipv4 != nil {
		network.IPv4Config = ipv4
	}
	if ipv6 != nil {
		network.IPv6Config = ipv6
	}
	err := m.writeConfig()
	if err != nil {
		return err
	}
	err = m.Load()
	return err
}
