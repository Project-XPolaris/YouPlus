package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

var Config AppConfig

type ShareFolderConfig struct {
	StorageId string `json:"storageId"`
	Part      string `json:"part"`
}
type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}
type AppConfig struct {
	Addr           string          `json:"addr"`
	ApiKey         string          `json:"api_key"`
	YouSMBAddr     string          `json:"yousmb_addr"`
	YouSMBRPC      string          `json:"yousmb_rpc"`
	Fstab          string          `json:"fstab"`
	NetConfig      string          `json:"net_config"`
	RPCAddr        string          `json:"rpc_addr"`
	DatabaseConfig *DatabaseConfig `json:"database"`
}

func LoadAppConfig() error {
	configFileName := os.Getenv("CONFIG_NAME")
	if len(configFileName) == 0 {
		configFileName = "./config.json"
	}
	logrus.Info("config from", configFileName)
	jsonFile, err := os.Open(configFileName)
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
