package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadJson(filePath string, target interface{}) error {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	raw, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(raw, target)
	return err
}
