package yousmb

import (
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/utils"
	"strings"
)

type CreateShareOption struct {
	Name       string
	Path       string
	Public     bool
	ValidUsers []string
	WriteList  []string
}

func CreateNewShare(option *CreateShareOption) error {
	properties := map[string]interface{}{
		"path":           option.Path,
		"browseable":     "yes",
		"available":      "yes",
		"directory mask": "0775",
		"create mask":    "0775",
		"writable":       "yes",
		"public":         "yes",
	}
	if !option.Public {
		properties["public"] = "no"
		properties["valid users"] = strings.Join(option.ValidUsers, ",")
		properties["write list"] = strings.Join(option.WriteList, ",")
	}
	requestBody := map[string]interface{}{
		"name":       option.Name,
		"properties": properties,
	}
	_, err := utils.POSTRequestWithJSON(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/add"), requestBody)
	if err != nil {
		return err
	}
	return nil
}

func AddUser(username string, password string) error {
	requestBody := map[string]interface{}{
		"username": username,
		"password": password,
	}
	_, err := utils.POSTRequestWithJSON(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/users"), requestBody)
	if err != nil {
		return err
	}
	return nil
}
