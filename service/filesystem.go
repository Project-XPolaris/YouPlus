package service

import (
	"github.com/projectxpolaris/youplus/database"
	"github.com/spf13/afero"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var DefaultFileSystem = &YouPlusFileSystem{
	entities: []FsEntity{},
}

func GetFileSystemWithUser(username string) (*YouPlusFileSystem, error) {
	userFileSystem := &YouPlusFileSystem{
		entities: []FsEntity{},
	}
	var user database.User
	err := database.Instance.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	userFolders, err := GetUserShareList(&user)
	for _, shareFolder := range userFolders {
		storageId := shareFolder.Folder.GetStorageId()
		if len(storageId) > 0 {
			userFileSystem.entities = append(userFileSystem.entities, FsEntity{
				Name:      shareFolder.Name,
				StorageId: storageId,
				Folder:    shareFolder,
			})
		}
	}
	userFileSystem.RootFile = &YouPlusFile{
		entities: userFileSystem.entities,
	}
	return userFileSystem, nil
}

func InitFileSystem() error {
	var shareFolders []database.ShareFolder
	err := database.Instance.Find(&shareFolders).Error
	if err != nil {
		return err
	}
	for _, shareFolder := range shareFolders {
		storageId := shareFolder.GetStorageId()
		if len(storageId) > 0 {
			DefaultFileSystem.entities = append(DefaultFileSystem.entities, FsEntity{
				Name:      shareFolder.Name,
				StorageId: storageId,
				Folder: &UserShareFolder{
					Name:   shareFolder.Name,
					Folder: &shareFolder,
					Access: true,
					Read:   true,
					Write:  true,
				},
			})
		}
	}
	DefaultFileSystem.RootFile = &YouPlusFile{
		entities: DefaultFileSystem.entities,
	}
	return nil
}

type FsFileInfo struct {
}

func (f *FsFileInfo) Name() string {
	return "Root"
}

func (f *FsFileInfo) Size() int64 {
	return 0
}

func (f *FsFileInfo) Mode() fs.FileMode {
	return os.ModePerm
}

func (f *FsFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *FsFileInfo) IsDir() bool {
	return true
}

func (f *FsFileInfo) Sys() any {
	return nil
}

type FsEntity struct {
	StorageId string
	Name      string
	Folder    *UserShareFolder
	FsEntity  *afero.BasePathFs
}
type EntityFileInfo struct {
	Entity FsEntity
}

func (f *EntityFileInfo) Name() string {
	return f.Entity.Name
}

func (f *EntityFileInfo) Size() int64 {
	return 0
}

func (f *EntityFileInfo) Mode() fs.FileMode {
	return os.ModePerm
}

func (f *EntityFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *EntityFileInfo) IsDir() bool {
	return true
}

func (f *EntityFileInfo) Sys() any {
	return nil
}

type YouPlusFileSystem struct {
	entities []FsEntity
	RootFile afero.File
}

func (f *YouPlusFileSystem) GetFs(path string) (*FsEntity, afero.Fs, error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	parts := strings.Split(path, string(os.PathSeparator))
	entityPath := parts[0]

	for _, entity := range f.entities {
		if entity.Name == entityPath {
			storageId := entity.StorageId
			storage := DefaultStoragePool.GetStorageById(storageId)
			return &entity, storage.GetFS(), nil
		}
	}
	return nil, nil, os.ErrNotExist
}
func (f *YouPlusFileSystem) Create(name string) (afero.File, error) {
	if isInRootDir(name) {
		return nil, os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return nil, err
	}
	if !entity.Folder.Access || entity.Folder.Write {
		return nil, os.ErrPermission
	}
	return fs.Create(name)
}

func (f *YouPlusFileSystem) Mkdir(name string, perm os.FileMode) error {
	if isInRootDir(name) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.Mkdir(name, perm)
}

func (f *YouPlusFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if isInRootDir(path) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(path)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.MkdirAll(path, perm)
}

func (f *YouPlusFileSystem) Open(name string) (afero.File, error) {
	if isInRootDir(name) {
		return f.RootFile, nil
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return nil, err
	}
	if !entity.Folder.Access || !entity.Folder.Read {
		return nil, os.ErrPermission
	}
	return fs.Open(name)
}

func (f *YouPlusFileSystem) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if isInRootDir(name) {
		return f.RootFile, nil
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return nil, err
	}
	if !entity.Folder.Access {
		return nil, os.ErrPermission
	}
	if flag&os.O_WRONLY != 0 {
		if !entity.Folder.Write {
			return nil, os.ErrPermission
		}
	}
	if flag&os.O_RDWR != 0 {
		if !entity.Folder.Write || !entity.Folder.Read {
			return nil, os.ErrPermission
		}
	}
	if flag&os.O_APPEND != 0 ||
		flag&os.O_TRUNC != 0 ||
		flag&os.O_CREATE != 0 ||
		flag&os.O_EXCL != 0 ||
		flag&os.O_SYNC != 0 {
		if !entity.Folder.Write {
			return nil, os.ErrPermission
		}
	}
	return fs.OpenFile(name, flag, perm)
}

func (f *YouPlusFileSystem) Remove(name string) error {
	if isInRootDir(name) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.Remove(name)
}

func (f *YouPlusFileSystem) RemoveAll(path string) error {
	if isInRootDir(path) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(path)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.RemoveAll(path)
}

func (f *YouPlusFileSystem) Rename(oldname, newname string) error {
	sourceEntity, oldFs, err := f.GetFs(oldname)
	if !sourceEntity.Folder.Access || !sourceEntity.Folder.Write {
		return os.ErrPermission
	}
	if err != nil {
		return err
	}
	targetEntity, newFs, err := f.GetFs(newname)
	if err != nil {
		return err
	}
	if !targetEntity.Folder.Access || !targetEntity.Folder.Write {
		return os.ErrPermission
	}
	if oldFs != newFs {
		return os.ErrPermission
	}

	return oldFs.Rename(oldname, newname)
}

func (f *YouPlusFileSystem) Stat(name string) (os.FileInfo, error) {
	if isInRootDir(name) {
		return f.RootFile.Stat()
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return nil, err
	}
	if !entity.Folder.Access || !entity.Folder.Read {
		return nil, os.ErrPermission
	}
	return fs.Stat(name)
}

func (f *YouPlusFileSystem) Name() string {
	return "YouPlus"
}

func (f *YouPlusFileSystem) Chmod(name string, mode os.FileMode) error {
	if isInRootDir(name) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.Chmod(name, mode)
}

func (f *YouPlusFileSystem) Chown(name string, uid, gid int) error {
	if isInRootDir(name) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.Chown(name, uid, gid)
}

func (f *YouPlusFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if isInRootDir(name) {
		return os.ErrPermission
	}
	entity, fs, err := f.GetFs(name)
	if err != nil {
		return err
	}
	if !entity.Folder.Access || !entity.Folder.Write {
		return os.ErrPermission
	}
	return fs.Chtimes(name, atime, mtime)
}
func (f *YouPlusFileSystem) GetRealPath(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	parts := strings.Split(path, string(os.PathSeparator))
	entityPath := parts[0]
	for _, entity := range f.entities {
		if entity.Name == entityPath {
			if !entity.Folder.Access {
				return "", os.ErrNotExist
			}
			storageId := entity.StorageId
			storage := DefaultStoragePool.GetStorageById(storageId)
			parts[0] = storage.GetRootPath()
			return filepath.Join(parts...), nil
		}
	}
	return "", os.ErrNotExist
}
func isInRootDir(target string) bool {
	if target == "." || target == "/" || target == "" {
		return true
	}
	return false
}
