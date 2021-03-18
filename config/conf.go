package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var Config AppConfig

type ShareFolderConfig struct {
	StorageId string `json:"storageId"`
	Part      string `json:"part"`
}
type StorageConfig struct {
	Id         string `json:"id"`
	Source     string `json:"source"`
	MountPoint string `json:"mount_point"`
}
type AppConfig struct {
	Addr       string               `json:"addr"`
	ApiKey     string               `json:"api_key"`
	YouSMBAddr string               `json:"yousmb_addr"`
	Folders    []*ShareFolderConfig `json:"folders"`
	Users      []string             `json:"users"`
	Fstab      string               `json:"fstab"`
	Storage    []*StorageConfig     `json:"storage"`
}

func LoadAppConfig() error {
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	raw, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(raw, &Config)
	return err
}
func (c *AppConfig) UpdateConfig() error {
	file, _ := json.MarshalIndent(c, "", "  ")
	err := ioutil.WriteFile("./config.json", file, 0644)
	return err
}
