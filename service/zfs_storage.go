package service

import (
	libzfs "github.com/bicomsystems/go-libzfs"
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"path"
)

type ZFSPoolStorage struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	MountPoint string `json:"mount_point"`
	PoolName   string
}

func (z *ZFSPoolStorage) GetName() string {
	return z.Name
}

func (z *ZFSPoolStorage) Update(option StorageUpdateOption) error {
	rawData := map[string]interface{}{}
	if option.Name != "" {
		rawData["Name"] = option.Name
	}
	err := database.Instance.Model(&database.ZFSStorage{}).Where("id = ?", z.Id).Updates(rawData).Error
	if err != nil {
		return err
	}
	if option.Name != "" {
		z.Name = option.Name
	}
	return nil
}

func (z *ZFSPoolStorage) SaveData() error {
	rawData := map[string]interface{}{}
	rawData["Name"] = z.Name
	rawData["MountPoint"] = z.MountPoint
	return database.Instance.Model(&database.ZFSStorage{}).Where("id = ?", z.Id).Updates(rawData).Error
}

func (z *ZFSPoolStorage) GetRootPath() string {
	return z.MountPoint
}

func (z *ZFSPoolStorage) GetId() string {
	return z.Id
}

func (z *ZFSPoolStorage) Remove() error {
	return database.Instance.Model(&database.ZFSStorage{}).Unscoped().Delete(&database.ZFSStorage{ID: z.Id}).Error
}

func (z *ZFSPoolStorage) GetUsage() (used int64, free int64, err error) {
	dataset, err := libzfs.DatasetOpen(z.MountPoint)
	if err != nil {
		return 0, 0, err
	}
	pool, err := dataset.Pool()
	if err != nil {
		return 0, 0, err
	}
	vtree, err := pool.VDevTree()
	if err != nil {
		return 0, 0, err
	}
	return int64(vtree.Stat.Alloc), int64(vtree.Stat.Space), nil

}

func (z *ZFSPoolStorage) LoadFromSave(data *database.ZFSStorage) error {
	dataset, err := libzfs.DatasetOpen(data.MountPoint)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	pool, err := dataset.Pool()
	if err != nil {
		return err
	}
	poolName, err := pool.Name()
	if err != nil {
		return err
	}
	z.Id = data.ID
	z.Name = data.Name
	z.MountPoint = data.MountPoint
	z.PoolName = poolName
	return nil
}

func CreateZFSStorage(datasetPath string) (Storage, error) {
	s := &ZFSPoolStorage{
		Id:         xid.New().String(),
		Name:       path.Base(datasetPath),
		MountPoint: datasetPath,
	}
	err := database.Instance.Save(&database.ZFSStorage{ID: s.Id, Name: s.Name, MountPoint: s.MountPoint}).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}
