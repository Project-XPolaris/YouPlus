package service

import (
	"errors"
	"github.com/mistifyio/go-zfs"
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"path"
)

var DefaultZFSManager = ZFSManager{}
var PoolNotFoundError = errors.New("target pool not found")

type ZFSManager struct {
	Pools []*zfs.Zpool
}

func (m *ZFSManager) LoadZFS() (err error) {
	m.Pools, err = zfs.ListZpools()
	return
}

func (m *ZFSManager) CreatePool(name string, paths ...string) error {
	_, err := zfs.CreateZpool(name, map[string]string{}, paths...)
	if err != nil {
		return err
	}
	// reload
	err = m.LoadZFS()
	if err != nil {
		return err
	}
	return nil
}
func (m *ZFSManager) GetPoolByName(name string) *zfs.Zpool {
	for _, pool := range m.Pools {
		if pool.Name == name {
			return pool
		}
	}
	return nil
}
func (m *ZFSManager) RemovePool(name string) error {
	pool := m.GetPoolByName(name)
	if pool == nil {
		return PoolNotFoundError
	}
	err := pool.Destroy()
	if err != nil {
		return err
	}
	// reload
	err = m.LoadZFS()
	if err != nil {
		return err
	}
	return nil
}

type ZFSPoolStorage struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	MountPoint string `json:"mount_point"`
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

func (z *ZFSPoolStorage) LoadFromSave(data *database.ZFSStorage) {
	z.Id = data.ID
	z.Name = data.Name
	z.MountPoint = data.MountPoint
}

func CreateZFSStorage(poolName string) (Storage, error) {
	s := &ZFSPoolStorage{
		Id:         xid.New().String(),
		Name:       path.Base(poolName),
		MountPoint: poolName,
	}
	err := database.Instance.Save(&database.ZFSStorage{ID: s.Id, Name: s.Name, MountPoint: s.MountPoint}).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}
