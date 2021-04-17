package application

import (
	libzfs "github.com/bicomsystems/go-libzfs"
	"github.com/projectxpolaris/youplus/service"
)

type AppTemplate struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Pid       int    `json:"pid"`
	Status    string `json:"status"`
	AutoStart bool   `json:"auto_start"`
	Icon      string `json:"icon"`
}

type ZFSPoolTemplate struct {
	Name      string `json:"name,omitempty"`
	Allocated uint64 `json:"allocated,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	Free      uint64 `json:"free,omitempty"`
}

func (t *ZFSPoolTemplate) Assign(pool libzfs.Pool) error {
	name, err := pool.Name()
	if err != nil {
		return err
	}
	t.Name = name
	return nil
}

type StorageTemplate struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func (t *StorageTemplate) Assign(storage service.Storage) {
	t.Id = storage.GetId()
	switch storage.(type) {
	case *service.DiskPartStorage:
		t.Type = "Parted"
	case *service.ZFSPoolStorage:
		t.Type = "ZFSPool"
	}
}

type ShareFolderUsers struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}
type ShareFolderTemplate struct {
	Id             uint               `json:"id"`
	Name           string             `json:"name"`
	Storage        StorageTemplate    `json:"storage,omitempty"`
	ValidateUsers  []ShareFolderUsers `json:"validateUsers,omitempty"`
	WriteableUsers []ShareFolderUsers `json:"writeableUsers,omitempty"`
	Public         string             `json:"public"`
	Readonly       string             `json:"readonly"`
	Writable       string             `json:"writable"`
}

type UserTemplate struct {
	Name string `json:"name"`
	Uid  string `json:"uid"`
}
type UserGroupTemplate struct {
	Name  string         `json:"name"`
	Gid   string         `json:"gid"`
	Type  string         `json:"type"`
	Users []UserTemplate `json:"users,omitempty"`
}

func (t *UserGroupTemplate) Assign(group *service.SystemUserGroup) {
	t.Name = group.Name
	t.Gid = group.Gid
	if group.Name == service.SuperuserGroup {
		t.Type = "admin"
	} else {
		t.Type = "normal"
	}
}
