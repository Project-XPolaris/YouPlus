package service

import (
	libzfs "github.com/bicomsystems/go-libzfs"
	"strings"
)

func (m *ZFSManager) PathToZFSPath(path string) (string, error) {
	datasets, err := libzfs.DatasetOpenAll()
	if err != nil {
		return "", err
	}
	for _, dataset := range datasets {
		mountPointProp := dataset.Properties[libzfs.DatasetPropMountpoint]
		if strings.HasPrefix(path, mountPointProp.Value+"/") {
			return strings.Replace(path, mountPointProp.Value, dataset.PoolName(), 1), nil
		}
	}
	return "", nil
}
func (m *ZFSManager) GetDatasetByPath(path string) (*libzfs.Dataset, error) {
	datasetPath, err := m.PathToZFSPath(path)
	if err != nil {
		return nil, err
	}
	if len(datasetPath) == 0 {
		return nil, nil
	}
	dataset, err := libzfs.DatasetOpen(datasetPath)
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}
