package yousmb

import (
	"fmt"
	"youplus/config"
	"youplus/utils"
)

func CreateNewShare(name string, path string) error {
	requestBody := map[string]interface{}{
		"name": name,
		"properties": map[string]interface{}{
			"path":           path,
			"browseable":     "yes",
			"available":      "yes",
			"directory mask": "0775",
			"create mask":    "0775",
			"writable":       "yes",
			"public":         "yes",
		},
	}
	_, err := utils.POSTRequestWithJSON(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/add"), requestBody)
	if err != nil {
		return err
	}
	return nil
}
