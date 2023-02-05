package service

import (
	"context"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
	"io/fs"
	"os"
)

type YouPlusWebdavFileSystem struct {
	Fs *YouPlusFileSystem
}

func (f *YouPlusWebdavFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return f.Fs.Mkdir(name, perm)
}

func (f *YouPlusWebdavFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := f.Fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &YouPlusWebdavFile{File: file}, nil
}

func (f *YouPlusWebdavFileSystem) RemoveAll(ctx context.Context, name string) error {
	return f.Fs.RemoveAll(name)
}

func (f *YouPlusWebdavFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return f.Fs.Rename(oldName, newName)
}

func (f *YouPlusWebdavFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return f.Fs.Stat(name)
}

type YouPlusWebdavFile struct {
	File afero.File
}

func (f *YouPlusWebdavFile) Close() error {
	return f.File.Close()
}

func (f *YouPlusWebdavFile) Read(p []byte) (n int, err error) {
	return f.File.Read(p)
}

func (f *YouPlusWebdavFile) Seek(offset int64, whence int) (int64, error) {
	return f.File.Seek(offset, whence)
}

func (f *YouPlusWebdavFile) Readdir(count int) ([]fs.FileInfo, error) {
	return f.File.Readdir(count)
}

func (f *YouPlusWebdavFile) Stat() (fs.FileInfo, error) {
	return f.File.Stat()
}

func (f *YouPlusWebdavFile) Write(p []byte) (n int, err error) {
	return f.File.Write(p)
}
