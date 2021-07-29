package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func WriteYaml(data interface{}, targetPath string) error {
	file, _ := yaml.Marshal(data)
	err := ioutil.WriteFile(targetPath, file, 0644)
	return err
}
