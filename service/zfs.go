package service

import (
	"errors"
	"fmt"
	"github.com/mistifyio/go-zfs"
	"github.com/rs/xid"
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

func (z *ZFSPoolStorage) GetRootPath() string {
	return z.MountPoint
}

func (z *ZFSPoolStorage) GetId() string {
	return z.Id
}

func (z *ZFSPoolStorage) Remove() error {
	// noting
	return nil
}

func (z *ZFSPoolStorage) SerializeSaveData() map[string]interface{} {
	rawData := map[string]interface{}{}
	rawData["id"] = z.Id
	rawData["name"] = z.Name
	rawData["mountPoint"] = z.MountPoint
	return rawData
}

func (z *ZFSPoolStorage) LoadFromSave(raw map[string]interface{}) {
	z.Id = raw["id"].(string)
	z.Name = raw["name"].(string)
	z.MountPoint = raw["mountPoint"].(string)
}

func CreateZFSStorage(poolName string) Storage {
	return &ZFSPoolStorage{
		Id:         xid.New().String(),
		Name:       poolName,
		MountPoint: fmt.Sprintf("/%s", poolName),
	}
}
