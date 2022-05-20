package service

import (
	"errors"
	"fmt"
	libzfs "github.com/bicomsystems/go-libzfs"
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

func (m *ZFSManager) GetDatasetList() ([]libzfs.Dataset, error) {
	return libzfs.DatasetOpenAll()
}
func (m *ZFSManager) CloseAllDataset(datasets []libzfs.Dataset) {
	libzfs.DatasetCloseAll(datasets)
}
func (m *ZFSManager) CreateDataset(datasetPath string) (dataset libzfs.Dataset, err error) {
	return libzfs.DatasetCreate(datasetPath, libzfs.DatasetTypeFilesystem, nil)
}

func (m *ZFSManager) DeleteDataset(datasetPath string) error {
	dataset, err := libzfs.DatasetOpen(datasetPath)
	if err != nil {
		return err
	}
	err = dataset.DestroyRecursive()
	if err != nil {
		return err
	}
	return nil
}

func (m *ZFSManager) CreateSnapshot(datasetPath string, snapshotName string) (libzfs.Dataset, error) {
	return libzfs.DatasetSnapshot(fmt.Sprintf("%s@%s", datasetPath, snapshotName), false, nil)
}

func (m *ZFSManager) GetDatasetSnapshotList(datasetPath string) ([]libzfs.Dataset, error) {
	dataset, err := libzfs.DatasetOpen(datasetPath)
	if err != nil {
		return nil, err
	}
	return dataset.Snapshots()
}

func (m *ZFSManager) DatasetRollback(datasetPath string, snapshotName string) error {
	dataset, err := libzfs.DatasetOpen(datasetPath)
	if err != nil {
		return err
	}
	snapshot, err := libzfs.DatasetOpen(fmt.Sprintf("%s@%s", datasetPath, snapshotName))
	if err != nil {
		return err
	}
	return dataset.Rollback(&snapshot, true)
}

func (m *ZFSManager) DeleteSnapshot(datasetPath string, snapshotName string) error {
	dataset, err := libzfs.DatasetOpen(fmt.Sprintf("%s@%s", datasetPath, snapshotName))
	if err != nil {
		return err
	}
	return dataset.Destroy(true)
}

type DatasetQueryFilter struct {
	Pool string `hsource:"query" hname:"pool"`
}

func (f *DatasetQueryFilter) isValid(dataset libzfs.Dataset) bool {
	if f.Pool != "" && f.Pool != dataset.PoolName() {
		return false
	}
	return true
}
func (m *ZFSManager) GetAllDataset(filter DatasetQueryFilter) ([]libzfs.Dataset, error) {
	queue, err := libzfs.DatasetOpenAll()
	if err != nil {
		return nil, err
	}
	result := make([]libzfs.Dataset, 0)
	for len(queue) != 0 {
		dataset := queue[0]
		if !dataset.IsSnapshot() {
			if dataset.Children != nil {
				queue = append(queue, dataset.Children...)
			}
			if filter.isValid(dataset) {
				result = append(result, dataset)
			}
		}
		if len(queue) > 0 {
			queue = queue[1:]
		} else {
			queue = []libzfs.Dataset{}
		}
	}
	return result, nil
}

type ZFSPoolListFilter struct {
	Name  string   `hsource:"query" hname:"name"`
	Disks []string `hsource:"query" hname:"disks"`
}

func CheckDiskInVdevTree(disk string, pool libzfs.Pool) (bool, error) {
	// walk the tree
	vt, err := pool.VDevTree()
	if err != nil {
		return false, err
	}
	queue := make([]libzfs.VDevTree, 0)
	queue = append(queue, vt)
	for len(queue) > 0 {
		curVt := queue[0]
		queue = queue[1:]
		if curVt.Name == disk {
			return true, nil
		}
		queue = append(queue, curVt.Devices...)
		queue = append(queue, curVt.Spares...)
		queue = append(queue, curVt.L2Cache...)
	}
	return false, err
}
func (m *ZFSManager) GetPoolList(filter *ZFSPoolListFilter) ([]libzfs.Pool, error) {
	result := make([]libzfs.Pool, 0)
	pools, err := libzfs.PoolOpenAll()
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		if len(filter.Name) > 0 {
			name, err := pool.Name()
			if err != nil {
				return nil, err
			}
			if name != filter.Name {
				continue
			}
		}
		if len(filter.Disks) > 0 {
			isExist := false
			for _, disk := range filter.Disks {
				exist, err := CheckDiskInVdevTree(disk, pool)
				if err != nil {
					return nil, err
				}
				if exist {
					isExist = true
					break
				}
			}
			if !isExist {
				continue
			}
		}
		result = append(result, pool)
	}
	return result, nil
}

func (m *ZFSManager) GetPoolByName(name string) (*libzfs.Pool, error) {
	pools, err := libzfs.PoolOpenAll()
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		poolName, err := pool.Name()
		if err != nil {
			return nil, err
		}
		if name == poolName {
			return &pool, nil
		}
	}
	return nil, nil
}
