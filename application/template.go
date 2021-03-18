package application

import (
	"github.com/mistifyio/go-zfs"
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

func (t *ZFSPoolTemplate) Assign(zpool *zfs.Zpool) {
	t.Name = zpool.Name
	t.Allocated = zpool.Allocated
	t.Size = zpool.Size
	t.Free = zpool.Free
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
