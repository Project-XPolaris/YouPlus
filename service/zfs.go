package service

import (
	"errors"
	libzfs "github.com/bicomsystems/go-libzfs"
	"github.com/projectxpolaris/youplus/database"
	"github.com/rs/xid"
	"path"
)

var DefaultZFSManager = ZFSManager{}
var PoolNotFoundError = errors.New("target pool not found")

type ZFSManager struct {
}

var VdevTypeMapping = map[string]libzfs.VDevType{
	"disk":   libzfs.VDevTypeDisk,
	"mirror": libzfs.VDevTypeMirror,
	"raidz":  libzfs.VDevTypeRaidz,
}

type Node struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Devices []Node `json:"devices"`
	Spares  []Node `json:"spares"`
	L2      []Node `json:"l2"`
}

func ConvertNodeToVDevTree(node *Node, vdev *libzfs.VDevTree) {
	vdev.Devices = []libzfs.VDevTree{}
	for _, dev := range node.Devices {
		devVdev := &libzfs.VDevTree{
			Type: VdevTypeMapping[dev.Type],
			Path: dev.Path,
		}
		ConvertNodeToVDevTree(&dev, devVdev)
		vdev.Devices = append(vdev.Devices, *devVdev)
	}
	vdev.Spares = []libzfs.VDevTree{}
	for _, dev := range node.Spares {
		devVdev := &libzfs.VDevTree{
			Type: VdevTypeMapping[dev.Type],
			Path: dev.Path,
		}
		ConvertNodeToVDevTree(&dev, devVdev)
		vdev.Spares = append(vdev.Spares, *devVdev)
	}
	vdev.L2Cache = []libzfs.VDevTree{}
	for _, dev := range node.Spares {
		devVdev := &libzfs.VDevTree{
			Type: VdevTypeMapping[dev.Type],
			Path: dev.Path,
		}
		ConvertNodeToVDevTree(&dev, devVdev)
		vdev.Spares = append(vdev.L2Cache, *devVdev)
	}
}
func (m *ZFSManager) CreatePoolWithNode(name string, rootNode Node) error {
	rootVdev := libzfs.VDevTree{}
	ConvertNodeToVDevTree(&rootNode, &rootVdev)
	return m.CreatePool(name, rootVdev)
}
func (m *ZFSManager) CreateSimpleDiskPool(name string, paths ...string) error {
	var vdev libzfs.VDevTree
	var mdevs []libzfs.VDevTree
	// build mirror devices specs
	for _, d := range paths {
		mdevs = append(mdevs, libzfs.VDevTree{Type: libzfs.VDevTypeDisk, Path: d})
	}
	// spare device specs
	// pool specs
	vdev.Devices = mdevs
	err := m.CreatePool(name, vdev)
	return err
}
func (m *ZFSManager) CreatePool(name string, vdev libzfs.VDevTree) error {
	// pool properties
	props := make(map[libzfs.Prop]string)
	// root dataset filesystem properties
	fsprops := make(map[libzfs.Prop]string)
	//err := os.MkdirAll("/" + name,os.ModePerm)
	//if err != nil {
	//	return err
	//}
	fsprops[libzfs.DatasetPropMountpoint] = "/" + name
	// pool features
	features := make(map[string]string)
	pool, err := libzfs.PoolCreate(name, vdev, features, props, fsprops)
	if err != nil {
		return err
	}
	pool.Close()
	dss, err := libzfs.DatasetOpenAll()

	for _, dataset := range dss {
		if dataset.PoolName() == name {
			dataset.Mount("", 0)
		}
	}
	return nil
}
func (m *ZFSManager) GetPoolCount() (int, error) {
	pools, err := libzfs.PoolOpenAll()
	if err != nil {
		return 0, err
	}
	defer libzfs.PoolCloseAll(pools)
	return len(pools), nil
}
func (m *ZFSManager) RemovePool(name string) error {
	pool, err := libzfs.PoolOpen(name)
	if err != nil {
		return err
	}
	ds, err := libzfs.DatasetOpen(name)
	if err != nil {
		return err
	}
	err = ds.Unmount(0)
	if err != nil {
		return err
	}
	defer ds.Close()
	defer pool.Close()
	err = pool.Destroy(name)
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
