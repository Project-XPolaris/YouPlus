package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var Config AppConfig

type AppConfig struct {
	Addr string `json:"addr"`
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
