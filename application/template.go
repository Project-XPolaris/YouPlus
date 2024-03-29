package application

import (
	libzfs "github.com/bicomsystems/go-libzfs"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
)

const TimeLayout = "2006-01-02 15:04:05"

type ZFSPoolTemplate struct {
	Name string          `json:"name,omitempty"`
	Tree ZFSTreeTemplate `json:"tree,omitempty"`
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
	Id       string                   `json:"id"`
	Name     string                   `json:"name"`
	Type     string                   `json:"type"`
	Used     int64                    `json:"used"`
	Total    int64                    `json:"total"`
	ZFS      *StorageZFSTemplate      `json:"zfs,omitempty"`
	DiskPart *StorageDiskPartTemplate `json:"diskPart,omitempty"`
	Path     *StoragePathTemplate     `json:"path,omitempty"`
}
type StorageZFSTemplate struct {
	Name string `json:"name"`
}
type StorageDiskPartTemplate struct {
	Name string `json:"name"`
}
type StoragePathTemplate struct {
	Path string `json:"path"`
}

func (t *StorageTemplate) Assign(storage service.Storage) {
	t.Id = storage.GetId()
	t.Name = storage.GetName()
	switch storage.(type) {
	case *service.DiskPartStorage:
		t.Type = "Parted"
		diskPartStorage := storage.(*service.DiskPartStorage)
		t.DiskPart = &StorageDiskPartTemplate{}
		t.DiskPart.Name = diskPartStorage.Source
	case *service.ZFSPoolStorage:
		t.Type = "ZFSPool"
		zfsStorage := storage.(*service.ZFSPoolStorage)
		t.ZFS = &StorageZFSTemplate{}
		t.ZFS.Name = zfsStorage.PoolName
	case *service.PathStorage:
		t.Type = "Path"
		pathStorage := storage.(*service.PathStorage)
		t.Path = &StoragePathTemplate{
			Path: pathStorage.Path,
		}
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
	ReadUsers     []ShareFolderUsers      `json:"readUsers"`
	WriteUsers    []ShareFolderUsers      `json:"writeUsers"`
	ValidGroups   []*UserGroupTemplate    `json:"validGroups"`
	InvalidGroups []*UserGroupTemplate    `json:"invalidGroups"`
	ReadGroups    []*UserGroupTemplate    `json:"readGroups"`
	WriteGroups   []*UserGroupTemplate    `json:"writeGroups"`
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

type Props struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Source string `json:"source"`
}
type DatasetTemplate struct {
	Pool          string  `json:"pool"`
	Path          string  `json:"path"`
	SnapshotCount int     `json:"snapshotCount,omitempty"`
	Props         []Props `json:"props,omitempty"`
}

func (t *DatasetTemplate) Assign(dataset *libzfs.Dataset) {
	t.Pool = dataset.PoolName()
	t.Path, _ = dataset.Path()
	t.Props = make([]Props, 0)
	for prop, property := range dataset.Properties {
		t.Props = append(t.Props, Props{
			Name:   libzfs.DatasetPropertyToName(prop),
			Value:  property.Value,
			Source: property.Source,
		})
	}
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

type SMBProcessStatusTemplate struct {
	PID      string `json:"pid"`
	Username string `json:"username"`
	Group    string `json:"group"`
	Machine  string `json:"machine"`
}

type SMBSharesStatusTemplate struct {
	Service   string `json:"service"`
	PID       string `json:"pid"`
	Machine   string `json:"machine"`
	ConnectAt string `json:"connectAt"`
}
