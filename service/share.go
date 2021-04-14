package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/projectxpolaris/youplus/yousmb"
	"gorm.io/gorm"
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
	err = yousmb.DefaultClient.CreateNewShare(&yousmb.CreateShareOption{
		Name:       option.Name,
		Path:       shareFolderPath,
		Public:     option.Public,
		ValidUsers: option.ValidUsers,
		WriteList:  option.WriteList,
	})
	if err != nil {
		return err
	}
	shareFolder := database.ShareFolder{
		Name: option.Name,
	}
	switch storage.(type) {
	case *ZFSPoolStorage:
		shareFolder.ZFSStorageId = storage.GetId()
	case *DiskPartStorage:
		shareFolder.PartStorageId = storage.GetId()
	}
	err = database.Instance.Save(&shareFolder).Error
	if err != nil {
		return err
	}
	return nil
}

func GetShareFolders() ([]*database.ShareFolder, error) {
	var folders []*database.ShareFolder
	err := database.Instance.Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}

type UpdateShareFolderOption struct {
	Name       string   `json:"name"`
	ValidUsers []string `json:"validUsers"`
	WriteList  []string `json:"writeList"`
	Public     string   `json:"public"`
	Readonly   string   `json:"readonly"`
	Writable   string   `json:"writable"`
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
	if len(option.Readonly) > 0 {
		requestBody.Properties["read only"] = option.Readonly
	}
	if len(option.Writable) > 0 {
		requestBody.Properties["writable"] = option.Writable
	}
	_, err := utils.POSTRequestWithJSON(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/update"), requestBody)
	return err
}

func RemoveShare(id uint) error {
	shareFolder := database.ShareFolder{
		Model: gorm.Model{ID: id},
	}
	err := database.Instance.Find(&shareFolder).Error
	if err != nil {
		return err
	}
	err = yousmb.DefaultClient.RemoveFolder(shareFolder.Name)
	if err != nil {
		return err
	}
	err = database.Instance.Model(&database.ShareFolder{}).
		Unscoped().
		Delete(&database.ShareFolder{Model: gorm.Model{ID: shareFolder.ID}}).
		Error
	if err != nil {
		return err
	}
	return nil
}
