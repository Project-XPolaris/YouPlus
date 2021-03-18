package service

import (
	"errors"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/yousmb"
	"os"
	"path/filepath"
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
