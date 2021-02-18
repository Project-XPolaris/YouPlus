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
	PartName   string   `json:"part_name,omitempty"`
	Name       string   `json:"name,omitempty"`
	Public     bool     `json:"public"`
	ValidUsers []string `json:"valid_users"`
	WriteList  []string `json:"write_list"`
}

func CreateNewShareFolder(option *NewShareFolderOption) error {
	part := GetPartByName(option.PartName)
	if part == nil {
		return PartNotFoundError
	}
	if len(part.MountPoint) == 0 {
		return PartNotMountError
	}
	shareFolderPath := filepath.Join(part.MountPoint, option.Name)
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
		PartName: option.PartName,
		Part:     option.Name,
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
