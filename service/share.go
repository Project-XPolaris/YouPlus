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
	StorageId string `json:"storageId,omitempty"`
	Name      string `json:"name,omitempty"`
	Public    bool   `json:"public,omitempty"`
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
		Name:   option.Name,
		Public: option.Public,
		Path:   shareFolderPath,
		Enable: false,
	}
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
func getSMBUserList(list []*database.User) []string {
	userList := make([]string, 0)
	if list == nil {
		return userList
	}
	for _, user := range list {
		userList = append(userList, user.Username)
	}
	return userList
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
		"valid users":    strings.Join(getSMBUserList(folder.ValidUsers), ","),
		"invalid users":  strings.Join(getSMBUserList(folder.InvalidUsers), ","),
		"read list":      strings.Join(getSMBUserList(folder.ReadUsers), ","),
		"write list":     strings.Join(getSMBUserList(folder.WriteUsers), ","),
	}
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
	err := database.Instance.Preload("ValidUsers").Preload("InvalidUsers").Preload("ReadUsers").Preload("WriteUsers").Find(&folders).Error
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
	Name         string   `json:"name"`
	ReadUsers    []string `json:"readUsers"`
	WriteUsers   []string `json:"writeUsers"`
	ValidUsers   []string `json:"validUsers"`
	InvalidUsers []string `json:"invalidUsers"`
	Public       bool     `json:"public"`
	Readonly     bool     `json:"readonly"`
	Enable       bool     `json:"enable"`
}

func putFolderUserList(folder *database.ShareFolder, usernameList []string, rel string) error {
	var users []*database.User
	err := database.Instance.Where("username in ?", usernameList).Find(&users).Error
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Clear()
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Append(users)
	if err != nil {
		return err
	}
	return nil
}
func UpdateSMBConfig(option *UpdateShareFolderOption) error {
	var folder database.ShareFolder
	err := database.Instance.Where("name = ?", option.Name).
		Preload("ValidUsers").
		Preload("InvalidUsers").
		Preload("ReadUsers").
		Preload("WriteUsers").
		First(&folder).
		Error
	if err != nil {
		return err
	}
	if option.ValidUsers != nil {
		err = putFolderUserList(&folder, option.ValidUsers, "ValidUsers")
		if err != nil {
			return err
		}
	}
	if option.InvalidUsers != nil {
		err = putFolderUserList(&folder, option.InvalidUsers, "InvalidUsers")
		if err != nil {
			return err
		}
	}
	if option.ReadUsers != nil {
		err = putFolderUserList(&folder, option.ReadUsers, "ReadUsers")
		if err != nil {
			return err
		}
	}
	if option.WriteUsers != nil {
		err = putFolderUserList(&folder, option.WriteUsers, "WriteUsers")
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
