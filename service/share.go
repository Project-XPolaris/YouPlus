package service

import (
	"errors"
	"os"
	"path/filepath"
	"youplus/config"
	"youplus/yousmb"
)

var (
	PartNotFoundError = errors.New("target part name not found")
	PartNotMountError = errors.New("target part not mounted")
)

type NewShareFolderOption struct {
	PartName string `json:"part_name,omitempty"`
	Name     string `json:"name,omitempty"`
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
	err = yousmb.CreateNewShare(option.Name, shareFolderPath)
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
