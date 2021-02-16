package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var Config AppConfig

type ShareFolderConfig struct {
	PartName string `json:"part_name,omitempty"`
	Part     string `json:"part,omitempty"`
}
type AppConfig struct {
	Addr       string               `json:"addr"`
	YouSMBAddr string               `json:"yousmb_addr"`
	Folders    []*ShareFolderConfig `json:"folders"`
	Users      []string             `json:"users"`
}

func LoadAppConfig() error {
	jsonFile, err := os.Open("config.json")
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
	err := ioutil.WriteFile("config.json", file, 0644)
	return err
}
