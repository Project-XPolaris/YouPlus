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

type AppConfig struct {
	Addr       string `json:"addr"`
	ApiKey     string `json:"api_key"`
	YouSMBAddr string `json:"yousmb_addr"`
	Fstab      string `json:"fstab"`
	NetConfig  string `json:"net_config"`
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
