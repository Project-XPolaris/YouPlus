package service

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"time"
)

type FolderFileSystem struct {
	FolderPath string
	fs         afero.Fs
}

func NewFolderFileSystem(folderPath string) *FolderFileSystem {
	return &FolderFileSystem{
		FolderPath: folderPath,
		fs:         afero.NewBasePathFs(afero.NewOsFs(), folderPath),
	}
}

func (s *FolderFileSystem) Create(name string) (afero.File, error) {
	return s.fs.Create(filepath.Join(s.FolderPath, name))
}

func (s *FolderFileSystem) Mkdir(name string, perm os.FileMode) error {
	return s.fs.Mkdir(filepath.Join(s.FolderPath, name), perm)
}

func (s *FolderFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return s.fs.MkdirAll(filepath.Join(s.FolderPath, path), perm)
}

func (s *FolderFileSystem) Open(name string) (afero.File, error) {
	return s.fs.Open(filepath.Join(s.FolderPath, name))
}

func (s *FolderFileSystem) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return s.fs.OpenFile(filepath.Join(s.FolderPath, name), flag, perm)
}

func (s *FolderFileSystem) Remove(name string) error {
	return s.fs.Remove(filepath.Join(s.FolderPath, name))
}

func (s *FolderFileSystem) RemoveAll(path string) error {
	return s.fs.RemoveAll(filepath.Join(s.FolderPath, path))
}

func (s *FolderFileSystem) Rename(oldname, newname string) error {
	return s.fs.Rename(filepath.Join(s.FolderPath, oldname), filepath.Join(s.FolderPath, newname))
}

func (s *FolderFileSystem) Stat(name string) (os.FileInfo, error) {
	return s.fs.Stat(filepath.Join(s.FolderPath, name))
}

func (s *FolderFileSystem) Name() string {
	return s.fs.Name()
}

func (s *FolderFileSystem) Chmod(name string, mode os.FileMode) error {
	return s.fs.Chmod(filepath.Join(s.FolderPath, name), mode)
}

func (s *FolderFileSystem) Chown(name string, uid, gid int) error {
	return s.fs.Chown(filepath.Join(s.FolderPath, name), uid, gid)
}

func (s *FolderFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.fs.Chtimes(filepath.Join(s.FolderPath, name), atime, mtime)
}
