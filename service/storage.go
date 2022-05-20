package service

import (
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/database"
	"github.com/sirupsen/logrus"
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
		err = s.LoadFromSave(zfsStorage)
		if err != nil {
			logrus.Error(err)
			continue
		}
		p.Storages = append(p.Storages, s)
	}
	var FolderStorageList []*database.FolderStorage
	err = database.Instance.Find(&FolderStorageList).Error
	if err != nil {
		return err
	}
	for _, folderStorage := range FolderStorageList {
		s := &PathStorage{}
		err = s.LoadFromSave(folderStorage)
		if err != nil {
			logrus.Error(err)
			continue
		}
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
	if storageType == "Path" {
		storage, err := CreatePathStorage(source)
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
func (p *StoragePool) UpdateStorage(id string, option StorageUpdateOption) error {
	storage := p.GetStorageById(id)
	if storage == nil {
		return StorageNotFoundError
	}
	err := storage.Update(option)
	if err != nil {
		return err
	}
	return nil
}

type StorageUpdateOption struct {
	Name string `json:"name"`
}
type Storage interface {
	GetId() string
	Remove() error
	SaveData() error
	GetRootPath() string
	GetName() string
	GetUsage() (used int64, free int64, err error)
	Update(option StorageUpdateOption) error
}
