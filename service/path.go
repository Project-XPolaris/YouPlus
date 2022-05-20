package service

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youplus/database"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

var PathNotFoundError = errors.New("target path not found")
var InvalidatePathError = errors.New("invalidate path")
var PathNotAccessible = errors.New("not accessible path")
var AddressConverterLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "AddressConverter",
})
var DefaultAddressConverterManager = AddressConverterManager{}

type Entity struct {
	Name        string
	Path        string
	Public      bool
	AccessUsers map[string]bool
}
type AddressConverterManager struct {
	Entities []*Entity
}
type PathItem struct {
	Name     string `json:"name"`
	RealPath string `json:"realPath"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

func (m *AddressConverterManager) Load() error {
	m.Entities = []*Entity{}
	var folders []database.ShareFolder
	err := database.Instance.
		Preload("ZFSStorage").
		Preload("PartStorage").
		Preload("ReadUsers").
		Where("enable = ?", true).
		Find(&folders).
		Error
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if folder.PartStorage == nil && folder.ZFSStorage == nil {
			continue
		}
		entity := &Entity{
			Name: folder.Name,
		}
		if folder.PartStorage != nil {
			entity.Path = filepath.Join(folder.PartStorage.MountPoint, folder.Name)
		}
		if folder.ZFSStorage != nil {
			entity.Path = filepath.Join(folder.ZFSStorage.MountPoint, folder.Name)
		}
		if folder.Public {
			entity.Public = true
		} else {
			entity.AccessUsers = map[string]bool{}
			for _, user := range folder.ReadUsers {
				entity.AccessUsers[user.Username] = true
			}
		}
		m.Entities = append(m.Entities, entity)
	}
	AddressConverterLogger.Info(fmt.Sprintf("success load %d entites", len(m.Entities)))
	return nil
}
func (m *AddressConverterManager) loadPermission(folders []database.ShareFolder) error {

	return nil
}
func (m *AddressConverterManager) ReadDir(target string, username string) ([]PathItem, error) {
	result := make([]PathItem, 0)
	// root path
	if target == "/" || len(target) == 0 {
		for _, entity := range m.Entities {
			if !entity.Public {
				if _, exist := entity.AccessUsers[username]; !exist {
					continue
				}
			}
			result = append(result, PathItem{
				Name:     entity.Name,
				RealPath: entity.Path,
				Path:     filepath.Join(entity.Name),
				Type:     "Directory",
			})
		}
		return result, nil
	}
	if strings.HasPrefix(target, "/") {
		target = target[1:]
	}
	pathParts := strings.Split(target, "/")
	rootDir := pathParts[0]
	entity := linq.From(m.Entities).FirstWith(func(i interface{}) bool {
		return i.(*Entity).Name == rootDir
	}).(*Entity)
	if entity == nil {
		return nil, PathNotFoundError
	}
	if !entity.Public {
		if _, exist := entity.AccessUsers[username]; !exist {
			return nil, PathNotAccessible
		}
	}
	realPath := filepath.Join(entity.Path, filepath.Join(pathParts[1:]...))
	if !strings.HasPrefix(realPath, "/") {
		realPath = "/" + realPath
	}
	items, err := os.ReadDir(realPath)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		pathItem := PathItem{
			Name:     item.Name(),
			RealPath: filepath.Join(realPath, item.Name()),
			Path:     filepath.Join(target, item.Name()),
		}
		if item.IsDir() {
			pathItem.Type = "Directory"
		} else {
			pathItem.Type = "File"
		}
		result = append(result, pathItem)
	}
	return result, nil
}

func (m *AddressConverterManager) GetRealPath(target string, username string) (string, error) {
	if target == "/" || len(target) == 0 {
		return "", InvalidatePathError
	}
	if strings.HasPrefix(target, "/") {
		target = target[1:]
	}
	pathParts := strings.Split(target, "/")
	rootDir := pathParts[0]
	entity := linq.From(m.Entities).FirstWith(func(i interface{}) bool {
		return i.(*Entity).Name == rootDir
	}).(*Entity)
	if entity == nil {
		return "", PathNotFoundError
	}
	if !entity.Public {
		if _, exist := entity.AccessUsers[username]; !exist {
			return "", PathNotAccessible
		}
	}

	realPath := filepath.Join(entity.Path, filepath.Join(pathParts[1:]...))
	return realPath, nil
}
