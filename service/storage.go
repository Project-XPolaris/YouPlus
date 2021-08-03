package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var StorageNotFoundError = errors.New("target storage not found")
var DefaultStoragePool = StoragePool{[]Storage{}}
var StoragePoolLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "StorageManager",
})

type StoragePool struct {
	Storages []Storage
}

func (p *StoragePool) LoadStorage() error {
	var diskPartStorage []*database.PartStorage
	err := database.Instance.Find(&diskPartStorage).Error
	if err != nil {
		return err
	}
	for _, partStorage := range diskPartStorage {
		diskPartStorage := &DiskPartStorage{}
		diskPartStorage.LoadFromSave(partStorage)
		p.Storages = append(p.Storages, diskPartStorage)
	}
	var ZFSStorageList []*database.ZFSStorage
	err = database.Instance.Find(&ZFSStorageList).Error
	if err != nil {
		return err
	}
	for _, zfsStorage := range ZFSStorageList {
		s := &ZFSPoolStorage{}
		s.LoadFromSave(zfsStorage)
		p.Storages = append(p.Storages, s)
	}
	StoragePoolLogger.Info(fmt.Sprintf("success load %d storages", len(p.Storages)))
	return nil
}
func (p *StoragePool) SaveStorage() error {
	for _, storage := range p.Storages {
		storage.SaveData()
	}
	return nil
}

func (p *StoragePool) NewStorage(source string, storageType string) error {
	if storageType == "DiskPart" {
		storage, err := NewDiskPartStorage(source)
		if err != nil {
			return err
		}
		p.Storages = append(p.Storages, storage)
	}
	if storageType == "ZFSPool" {
		storage, err := CreateZFSStorage(source)
		if err != nil {
			return err
		}
		p.Storages = append(p.Storages, storage)
	}
	err := p.SaveStorage()
	if err != nil {
		return err
	}
	return nil
}

func (p *StoragePool) RemoveStorage(id string) error {
	var targetStorage Storage
	var targetIndex int
	for idx, storage := range p.Storages {
		if storage.GetId() == id {
			targetStorage = storage
			targetIndex = idx
			break
		}
	}
	if targetStorage == nil {
		return nil
	}
	err := targetStorage.Remove()
	if err != nil {
		return err
	}

	// update config
	p.Storages[targetIndex] = p.Storages[len(p.Storages)-1]
	p.Storages = p.Storages[0 : len(p.Storages)-1]
	err = p.SaveStorage()
	if err != nil {
		return err
	}
	return nil
}
func (p *StoragePool) GetStorageById(id string) Storage {
	for _, storage := range p.Storages {
		if storage.GetId() == id {
			return storage
		}
	}
	return nil
}

type Storage interface {
	GetId() string
	Remove() error
	SaveData() error
	GetRootPath() string
	GetUsage() (used int64, free int64, err error)
}

type DiskPartStorage struct {
	Id         string `json:"id"`
	Source     string `json:"source"`
	MountPoint string `json:"mount_point"`
}

func (s *DiskPartStorage) GetRootPath() string {
	return s.MountPoint
}

func (s *DiskPartStorage) LoadFromSave(data *database.PartStorage) {
	s.Id = data.ID
	s.Source = data.Source
	s.MountPoint = data.MountPoint
}

func (s *DiskPartStorage) SaveData() error {
	rawData := map[string]interface{}{}
	rawData["source"] = s.Source
	rawData["mountPoint"] = s.MountPoint
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
