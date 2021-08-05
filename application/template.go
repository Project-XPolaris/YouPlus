package application

import (
	libzfs "github.com/bicomsystems/go-libzfs"
	"github.com/projectxpolaris/youplus/database"
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
	Name string          `json:"name,omitempty"`
	Tree ZFSTreeTemplate `json:"tree"`
}

func (t *ZFSPoolTemplate) Assign(pool libzfs.Pool) error {
	name, err := pool.Name()
	if err != nil {
		return err
	}
	t.Name = name
	vt, err := pool.VDevTree()
	if err != nil {
		return err
	}
	if vt.Devices != nil {
		t.Tree = ZFSTreeTemplate{}
		t.Tree.Assign(&vt)
	}
	return nil
}

type ZFSTreeTemplate struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Allocated uint64            `json:"allocated,omitempty"`
	Size      uint64            `json:"size,omitempty"`
	Free      uint64            `json:"free,omitempty"`
	Alloc     uint64            `json:"alloc,omitempty"`
	Path      string            `json:"path"`
	Devices   []ZFSTreeTemplate `json:"devices"`
	L2Cache   []ZFSTreeTemplate `json:"l2Cache"`
	Spares    []ZFSTreeTemplate `json:"spares"`
}

func (t *ZFSTreeTemplate) Assign(tree *libzfs.VDevTree) {
	t.Name = tree.Name
	t.Type = string(tree.Type)
	t.Size = tree.Stat.Space
	t.Alloc = tree.Stat.Alloc
	t.Free = tree.Stat.Space - tree.Stat.Alloc
	t.Path = tree.Path
	t.Devices = []ZFSTreeTemplate{}
	if tree.Devices != nil {
		for _, device := range tree.Devices {
			template := ZFSTreeTemplate{}
			template.Assign(&device)
			t.Devices = append(t.Devices, template)
		}
	}
	t.L2Cache = []ZFSTreeTemplate{}
	if tree.L2Cache != nil {
		for _, l2 := range tree.L2Cache {
			template := ZFSTreeTemplate{}
			template.Assign(&l2)
			t.L2Cache = append(t.L2Cache, template)
		}
	}
	t.Spares = []ZFSTreeTemplate{}
	if tree.Spares != nil {
		for _, spare := range tree.Spares {
			template := ZFSTreeTemplate{}
			template.Assign(&spare)
			t.Spares = append(t.Spares, template)
		}
	}
}

type StorageTemplate struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Used  int64  `json:"used"`
	Total int64  `json:"total"`
}

func (t *StorageTemplate) Assign(storage service.Storage) {
	t.Id = storage.GetId()
	switch storage.(type) {
	case *service.DiskPartStorage:
		t.Type = "Parted"
	case *service.ZFSPoolStorage:
		t.Type = "ZFSPool"
	}
	t.Used, t.Total, _ = storage.GetUsage()
}

type ShareFolderUsers struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}
type ShareFolderTemplate struct {
	Id            uint                    `json:"id"`
	Name          string                  `json:"name"`
	Storage       StorageTemplate         `json:"storage,omitempty"`
	ValidUsers    []ShareFolderUsers      `json:"validUsers"`
	InvalidUsers  []ShareFolderUsers      `json:"invalidUsers"`
	ReadUsers     []ShareFolderUsers      `json:"readUsers,omitempty"`
	WriteUsers    []ShareFolderUsers      `json:"writeUsers,omitempty"`
	ValidGroups   []*UserGroupTemplate    `json:"validGroups"`
	InvalidGroups []*UserGroupTemplate    `json:"invalidGroups"`
	ReadGroups    []*UserGroupTemplate    `json:"readGroups,omitempty"`
	WriteGroups   []*UserGroupTemplate    `json:"writeGroups,omitempty"`
	Enable        bool                    `json:"enable"`
	Public        bool                    `json:"public"`
	Readonly      bool                    `json:"readonly"`
	Guest         service.UserShareFolder `json:"guest"`
	Other         service.UserShareFolder `json:"other"`
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
func SerializeGroups(groups []*database.UserGroup) []*UserGroupTemplate {
	data := make([]*UserGroupTemplate, 0)
	for _, group := range groups {
		systemGroup := service.DefaultUserManager.GetGroupById(group.Gid)
		if systemGroup == nil {
			continue
		}
		template := UserGroupTemplate{}
		template.Assign(systemGroup)
		data = append(data, &template)
	}
	return data
}

type DatasetTemplate struct {
	Pool          string `json:"pool"`
	Path          string `json:"path"`
	SnapshotCount int    `json:"snapshotCount,omitempty"`
}

func (t *DatasetTemplate) Assign(dataset *libzfs.Dataset) {
	t.Pool = dataset.PoolName()
	t.Path, _ = dataset.Path()
	snapshots, err := dataset.Snapshots()
	if err == nil {
		t.SnapshotCount = len(snapshots)
	}
}

func SerializerDatasetTemplates(datasets []libzfs.Dataset) []DatasetTemplate {
	data := make([]DatasetTemplate, 0)
	for _, dataset := range datasets {
		template := DatasetTemplate{}
		template.Assign(&dataset)
		data = append(data, template)
	}
	return data
}
