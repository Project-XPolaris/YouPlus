package service

import "github.com/projectxpolaris/youplus/utils"

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

func GetPartByName(name string) *Part {
	disks := utils.Lsblk()
	for _, block := range disks {
		if block["type"] == "part" && block["name"] == name {
			return newPartFromRaw(block)
		}
	}
	return nil
}
