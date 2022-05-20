package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"github.com/shirou/gopsutil/v3/disk"
	"os"
	"path/filepath"
)

type DiskPartStorage struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	MountPoint string `json:"mount_point"`
}

func (s *DiskPartStorage) GetName() string {
	return s.Name
}

func (s *DiskPartStorage) Update(option StorageUpdateOption) error {
	rawData := map[string]interface{}{}
	if option.Name != "" {
		rawData["source"] = option.Name
	}
	err := database.Instance.Model(&database.PartStorage{}).Where("id = ?", s.Id).Updates(rawData).Error
	if err != nil {
		return err
	}
	if option.Name != "" {
		s.Name = option.Name
	}
	return nil
}

func (s *DiskPartStorage) GetRootPath() string {
	return s.MountPoint
}

func (s *DiskPartStorage) LoadFromSave(data *database.PartStorage) {
	s.Id = data.ID
	s.Source = data.Source
	s.MountPoint = data.MountPoint
	s.Name = data.Name
}

func (s *DiskPartStorage) SaveData() error {
	rawData := map[string]interface{}{}
	rawData["source"] = s.Source
	rawData["mountPoint"] = s.MountPoint
	rawData["name"] = s.Name
	return database.Instance.Model(&database.PartStorage{}).Where("id = ?", s.Id).Updates(rawData).Error
}

func (s *DiskPartStorage) GetId() string {
	return s.Id
}

func (s *DiskPartStorage) Remove() error {
	// unmount
	err := DefaultFstab.RemoveMount(s.MountPoint)
	if err != nil {
		return err
	}
	err = DefaultFstab.Save()
	if err != nil {
		return err
	}
	err = DefaultFstab.Reload()
	if err != nil {
		return err
	}
	//clear mount folder
	err = os.Remove(s.MountPoint)
	if err != nil {
		return err
	}
	return nil
}
func (s *DiskPartStorage) GetUsage() (used int64, free int64, err error) {
	part := GetPartByName(s.Source)
	if part == nil {
		return 0, 0, errors.New("unknown fs type")
	}
	stat, err := disk.Usage(s.Source)
	if err != nil {
		return 0, 0, err
	}
	return int64(stat.Used), int64(stat.Total), err
}

func NewDiskPartStorage(source string) (Storage, error) {
	id := xid.New().String()
	storage := &DiskPartStorage{
		Id:         id,
		Source:     source,
		Name:       id,
		MountPoint: fmt.Sprintf(filepath.Join("mnt", id)),
	}
	//read fstype
	part := GetPartByName(filepath.Base(source))
	if part == nil {
		return nil, errors.New("unknown fs type")
	}
	//init mount dir
	err := os.MkdirAll(storage.MountPoint, os.ModePerm)
	if err != nil {
		return nil, err
	}
	option := &AddMountOption{
		Spec:    storage.Source,
		File:    storage.MountPoint,
		VfsType: part.FSType,
		MntOps: map[string]string{
			"defaults": "",
		},
		Freq:   0,
		PassNo: 0,
	}
	DefaultFstab.AddMount(option)
	err = DefaultFstab.Save()
	if err != nil {
		return nil, err
	}
	err = DefaultFstab.Reload()
	if err != nil {
		return nil, err
	}

	//save
	err = database.Instance.Save(&database.PartStorage{
		ID:           id,
		MountPoint:   storage.MountPoint,
		Name:         filepath.Base(storage.MountPoint),
		Source:       storage.Source,
		ShareFolders: nil,
	}).Error
	if err != nil {
		return nil, err
	}
	return storage, nil
}
