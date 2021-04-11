package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/projectxpolaris/youplus/yousmb"
	"os"
	"path/filepath"
	"strings"
)

var (
	PartNotFoundError = errors.New("target part name not found")
	PartNotMountError = errors.New("target part not mounted")
)

type NewShareFolderOption struct {
	StorageId  string   `json:"storageId,omitempty"`
	Name       string   `json:"name,omitempty"`
	Public     bool     `json:"public,omitempty"`
	ValidUsers []string `json:"valid_users,omitempty"`
	WriteList  []string `json:"write_list,omitempty"`
}

func CreateNewShareFolder(option *NewShareFolderOption) error {
	storage := DefaultStoragePool.GetStorageById(option.StorageId)
	if storage == nil {
		return StorageNotFoundError
	}
	shareFolderPath := filepath.Join(storage.GetRootPath(), option.Name)
	err := os.MkdirAll(shareFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	err = yousmb.CreateNewShare(&yousmb.CreateShareOption{
		Name:       option.Name,
		Path:       shareFolderPath,
		Public:     option.Public,
		ValidUsers: option.ValidUsers,
		WriteList:  option.WriteList,
	})
	if err != nil {
		return err
	}
	config.Config.Folders = append(config.Config.Folders, &config.ShareFolderConfig{
		StorageId: option.StorageId,
		Part:      option.Name,
	})
	err = config.Config.UpdateConfig()
	if err != nil {
		return err
	}
	return nil
}

func GetShareFolders() ([]*config.ShareFolderConfig, error) {
	return config.Config.Folders, nil
}

type UpdateShareFolderOption struct {
	Name       string   `json:"name"`
	ValidUsers []string `json:"validUsers"`
	WriteList  []string `json:"writeList"`
	Public     string   `json:"public"`
}
type SMBFolderRequestBody struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

func UpdateSMBConfig(option *UpdateShareFolderOption) error {
	requestBody := SMBFolderRequestBody{
		Name:       option.Name,
		Properties: map[string]interface{}{},
	}

	if option.ValidUsers != nil && len(option.ValidUsers) > 0 {
		for _, validUser := range option.ValidUsers {
			user := DefaultUserManager.GetUserByName(validUser)
			if user == nil {
				return errors.New("invalidate user")
			}

		}
		requestBody.Properties["valid users"] = strings.Join(option.ValidUsers, ",")
	}
	if option.WriteList != nil && len(option.WriteList) > 0 {
		for _, writeUser := range option.WriteList {
			user := DefaultUserManager.GetUserByName(writeUser)
			if user == nil {
				return errors.New("invalidate user")
			}

		}
		requestBody.Properties["write list"] = strings.Join(option.WriteList, ",")
	}
	if len(option.Public) > 0 {
		requestBody.Properties["public"] = option.Public
	}
	_, err := utils.POSTRequestWithJSON(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/update"), requestBody)
	return err
}
