package service

import (
	"errors"
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
	Readonly   bool     `json:"readonly"`
	ReadUsers  []string `json:"readUsers,omitempty"`
	WriteUsers []string `json:"writeUsers,omitempty"`
}

func CreateNewShareFolder(option *NewShareFolderOption) error {
	// get storage
	storage := DefaultStoragePool.GetStorageById(option.StorageId)
	if storage == nil {
		return StorageNotFoundError
	}
	// create share directory
	shareFolderPath := filepath.Join(storage.GetRootPath(), option.Name)
	err := os.MkdirAll(shareFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	// create share

	shareFolder := database.ShareFolder{
		Name:     option.Name,
		Public:   option.Public,
		Path:     shareFolderPath,
		Readonly: option.Readonly,
		Enable:   true,
	}
	var validUsers []*database.User
	err = database.Instance.Where("username in ?", option.ReadUsers).Find(&validUsers).Error
	if err != nil {
		return err
	}
	shareFolder.ReadUsers = validUsers
	var writeUsers []*database.User
	err = database.Instance.Where("username in ?", option.WriteUsers).Find(&writeUsers).Error
	if err != nil {
		return err
	}
	shareFolder.WriteUsers = writeUsers
	switch storage.(type) {
	case *ZFSPoolStorage:
		shareFolder.ZFSStorageId = storage.GetId()
	case *DiskPartStorage:
		shareFolder.PartStorageId = storage.GetId()
	}
	err = SyncShareFolderOptionToSMB(&shareFolder)
	if err != nil {
		return err
	}
	err = database.Instance.Save(&shareFolder).Error
	if err != nil {
		return err
	}
	return nil
}
func SyncShareFolderOptionToSMB(folder *database.ShareFolder) error {
	properties := map[string]interface{}{
		"path":           folder.Path,
		"create mask":    "0775",
		"directory mask": "0775",
		"read only":      utils.GetSmbBoolText(folder.Readonly),
		"available":      utils.GetSmbBoolText(folder.Enable),
		"browseable":     "yes",
		"public":         utils.GetSmbBoolText(folder.Public),
	}
	validUsers := []string{}
	for _, readUser := range folder.ReadUsers {
		validUsers = append(validUsers, readUser.Username)
	}
	properties["valid_users"] = strings.Join(validUsers, ",")
	writeList := []string{}
	for _, writeUser := range folder.WriteUsers {
		writeList = append(writeList, writeUser.Username)
	}
	properties["write_list"] = strings.Join(writeList, ",")
	if folder.Public {
		properties["public"] = "yes"
	}
	response, err := yousmb.DefaultClient.GetConfig()
	if err != nil {
		return err
	}
	for _, section := range response.Sections {
		// update
		if section.Name == folder.Name {
			err = yousmb.DefaultClient.UpdateFolder(&yousmb.FolderRequestBody{Name: folder.Name, Properties: properties})
			return err
		}
	}
	err = yousmb.DefaultClient.CreateNewShareWithRaw(map[string]interface{}{
		"name":       folder.Name,
		"properties": properties,
	})
	return err
}
func GetShareFolders() ([]*database.ShareFolder, error) {
	var folders []*database.ShareFolder
	err := database.Instance.Preload("ReadUsers").Preload("WriteUsers").Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}
func GetShareFolderCount() (int64, error) {
	var count int64
	err := database.Instance.Model(&database.ShareFolder{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

type UpdateShareFolderOption struct {
	Name       string   `json:"name"`
	ReadUsers  []string `json:"readUsers"`
	WriteUsers []string `json:"writeUsers"`
	Public     bool     `json:"public"`
	Readonly   bool     `json:"readonly"`
	Enable     bool     `json:"enable"`
}

func UpdateSMBConfig(option *UpdateShareFolderOption) error {
	var folder database.ShareFolder
	err := database.Instance.Where("name = ?", option.Name).Preload("ReadUsers").Preload("WriteUsers").First(&folder).Error
	if err != nil {
		return err
	}
	if option.ReadUsers != nil {
		var validUsers []*database.User
		err = database.Instance.Where("username in ?", option.ReadUsers).Find(&validUsers).Error
		if err != nil {
			return err
		}
		folder.ReadUsers = validUsers
		err = database.Instance.Model(&folder).Association("ReadUsers").Clear()
		if err != nil {
			return err
		}
		err = database.Instance.Model(&folder).Association("ReadUsers").Append(validUsers)
		if err != nil {
			return err
		}
	}
	if option.WriteUsers != nil {
		var writeUsers []*database.User
		err = database.Instance.Where("username in ?", option.WriteUsers).Find(&writeUsers).Error
		if err != nil {
			return err
		}
		folder.WriteUsers = writeUsers
		err = database.Instance.Model(&folder).Association("WriteUsers").Clear()
		if err != nil {
			return err
		}
		err = database.Instance.Model(&folder).Association("WriteUsers").Append(writeUsers)
		if err != nil {
			return err
		}
	}
	folder.Public = option.Public
	folder.Enable = option.Enable
	folder.Readonly = option.Readonly
	err = SyncShareFolderOptionToSMB(&folder)
	if err != nil {
		return err
	}
	err = database.Instance.Save(&folder).Error
	if err != nil {
		return err
	}

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
