package service

import (
	"encoding/json"
	"fmt"
	"github.com/projectxpolaris/youplus/utils"
	"os/exec"
	"strings"
)

type Disk struct {
	Name  string  `json:"name,omitempty"`
	Model string  `json:"model,omitempty"`
	Size  string  `json:"size,omitempty"`
	Parts []*Part `json:"parts,omitempty"`
}

type Part struct {
	Name       string `json:"name,omitempty"`
	FSType     string `json:"fs_type,omitempty"`
	Size       string `json:"size,omitempty"`
	MountPoint string `json:"mountpoint,omitempty"`
}

func newPartFromRaw(block map[string]string) *Part {
	return &Part{
		Name:       block["name"],
		FSType:     block["fstype"],
		Size:       block["size"],
		MountPoint: block["mountpoint"],
	}
}
func ReadDiskList() []*Disk {
	disks := utils.Lsblk()
	result := make([]*Disk, 0)
	for _, block := range disks {
		if block["type"] == "disk" {
			disk := &Disk{
				Name:  block["name"],
				Model: block["model"],
				Size:  block["size"],
				Parts: []*Part{},
			}
			result = append(result, disk)
		}
	}
	for _, block := range disks {
		if block["type"] == "part" {
			for _, disk := range result {
				if disk.Name == block["pkname"] {
					disk.Parts = append(disk.Parts, newPartFromRaw(block))
				}
			}
		}
	}
	return result
}
func GetDiskByName(name string) *Disk {
	disks := ReadDiskList()
	for _, disk := range disks {
		if disk.Name == name {
			return disk
		}
	}
	return nil
}
func GetPartByName(name string) *Part {
	disks := utils.Lsblk()
	for _, block := range disks {
		if block["type"] == "part" && block["name"] == name {
			return newPartFromRaw(block)
		}
	}
	return nil
}

type SmartInfoAttr struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Worst     int    `json:"worst"`
	Threshold int    `json:"threshold"`
	Value     int    `json:"value"`
}
type DiskSmartInfo struct {
	ModelFamily  string          `json:"modelFamily"`
	ModelName    string          `json:"modelName"`
	SerialNumber string          `json:"serialNumber"`
	Status       bool            `json:"status"`
	Attrs        []SmartInfoAttr `json:"attrs"`
}

func (d *Disk) GetSmartInfo() (map[string]interface{}, error) {
	if strings.HasPrefix(d.Name, "nvme") {
		return nil, nil
	}
	cmd := exec.Command("smartctl", "--all", "--json", fmt.Sprintf("/dev/%s", d.Name))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
