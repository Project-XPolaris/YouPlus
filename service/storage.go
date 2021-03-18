package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

var saveName = "./storage.json"
var StorageNotFoundError = errors.New("target storage not found")
var DefaultStoragePool = StoragePool{[]Storage{}}
var StoragePoolLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "StorageManager",
})

type StoragePool struct {
	Storages []Storage
}

func (p *StoragePool) LoadStorage() error {
	jsonFile, err := os.Open(saveName)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	raw, _ := ioutil.ReadAll(jsonFile)
	var rawConfig []map[string]interface{}
	err = json.Unmarshal(raw, &rawConfig)

	for _, rawStorage := range rawConfig {
		switch rawStorage["type"] {
		case "DiskPart":
			diskPartStorage := &DiskPartStorage{}
			diskPartStorage.LoadFromSave(rawStorage)
			p.Storages = append(p.Storages, diskPartStorage)
		case "ZFSPool":
			zfsStorage := &ZFSPoolStorage{}
			zfsStorage.LoadFromSave(rawStorage)
			p.Storages = append(p.Storages, zfsStorage)
		}

	}
	StoragePoolLogger.Info(fmt.Sprintf("success load %d storages", len(p.Storages)))
	return nil
}
func (p *StoragePool) SaveStorage() error {
	rawData := make([]map[string]interface{}, 0)
	for _, storage := range p.Storages {
		saveInfo := storage.SerializeSaveData()
		switch storage.(type) {
		case *DiskPartStorage:
			saveInfo["type"] = "DiskPart"
		case *ZFSPoolStorage:
			saveInfo["type"] = "ZFSPool"
		}
		rawData = append(rawData, saveInfo)
	}

	file, _ := json.MarshalIndent(rawData, "", "  ")
	err := ioutil.WriteFile(saveName, file, 0644)
	return err
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
		storage := CreateZFSStorage(source)
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
	SerializeSaveData() map[string]interface{}
	LoadFromSave(raw map[string]interface{})
	GetRootPath() string
}

type DiskPartStorage struct {
	Id         string `json:"id"`
	Source     string `json:"source"`
	MountPoint string `json:"mount_point"`
}

func (s *DiskPartStorage) GetRootPath() string {
	return s.MountPoint
}

func (s *DiskPartStorage) LoadFromSave(raw map[string]interface{}) {
	s.Id = raw["id"].(string)
	s.Source = raw["source"].(string)
	s.MountPoint = raw["mountPoint"].(string)
}

func (s *DiskPartStorage) SerializeSaveData() map[string]interface{} {
	rawData := map[string]interface{}{}
	rawData["id"] = s.Id
	rawData["source"] = s.Source
	rawData["mountPoint"] = s.MountPoint
	return rawData
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

func NewDiskPartStorage(source string) (Storage, error) {
	id := xid.New().String()
	storage := &DiskPartStorage{
		Id:         id,
		Source:     source,
		MountPoint: "/" + fmt.Sprintf(filepath.Join("mnt", id)),
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
	config.Config.Storage = append(config.Config.Storage, &config.StorageConfig{
		Id:         storage.Id,
		Source:     storage.Source,
		MountPoint: storage.MountPoint,
	})
	err = config.Config.UpdateConfig()
	if err != nil {
		return nil, err
	}
	return storage, nil
}
