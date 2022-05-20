package service

import (
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"github.com/shirou/gopsutil/disk"
	"path/filepath"
)

type PathStorage struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (s *PathStorage) GetId() string {
	return s.Id
}

func (s *PathStorage) Remove() error {
	return database.Instance.Model(&database.FolderStorage{}).Where("id = ?", s.Id).Delete(PathStorage{}).Error
}

func (s *PathStorage) SaveData() error {
	rawData := map[string]interface{}{}
	rawData["Name"] = s.Name
	rawData["Path"] = s.Path
	return database.Instance.Model(&database.FolderStorage{}).Where("id = ?", s.Id).Updates(rawData).Error
}

func (s *PathStorage) GetRootPath() string {
	return s.Path
}

func (s *PathStorage) GetName() string {
	return s.Name
}

func (s *PathStorage) GetUsage() (used int64, free int64, err error) {
	stat, err := disk.Usage(s.Path)
	return int64(stat.Used), int64(stat.Total), err
}

func (s *PathStorage) Update(option StorageUpdateOption) error {
	rawData := map[string]interface{}{}
	if option.Name != "" {
		rawData["Name"] = option.Name
	}
	err := database.Instance.Model(&database.FolderStorage{}).Where("id = ?", s.Id).Updates(rawData).Error
	if err != nil {
		return err
	}
	if option.Name != "" {
		s.Name = option.Name
	}
	return nil
}
func (s *PathStorage) LoadFromSave(storage *database.FolderStorage) error {
	s.Id = storage.ID
	s.Name = storage.Name
	s.Path = storage.Path
	return nil
}
func CreatePathStorage(path string) (*PathStorage, error) {
	var storagePath = &PathStorage{
		Id:   xid.New().String(),
		Name: filepath.Base(path),
		Path: path,
	}
	err := database.Instance.Save(&database.FolderStorage{ID: storagePath.Id, Name: storagePath.Name, Path: storagePath.Path}).Error
	if err != nil {
		return nil, err
	}
	return storagePath, nil
}
