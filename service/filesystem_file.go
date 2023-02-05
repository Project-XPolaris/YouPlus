package service

import "os"

type YouPlusFile struct {
	entities []FsEntity
}

func (f *YouPlusFile) Close() error {
	return nil
}

func (f *YouPlusFile) Read(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}

func (f *YouPlusFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrPermission
}

func (f *YouPlusFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrPermission
}

func (f *YouPlusFile) Write(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}

func (f *YouPlusFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrPermission
}

func (f *YouPlusFile) Name() string {
	return "YouPlus"
}

func (f *YouPlusFile) Readdir(count int) ([]os.FileInfo, error) {
	items := make([]os.FileInfo, 0)
	for _, entity := range f.entities {
		items = append(items, &EntityFileInfo{Entity: entity})
	}
	return items, nil
}

func (f *YouPlusFile) Readdirnames(n int) ([]string, error) {
	names := make([]string, 0)
	for _, entity := range f.entities {
		names = append(names, entity.Name)
	}
	return names, nil
}

func (f *YouPlusFile) Stat() (os.FileInfo, error) {
	return &FsFileInfo{}, nil
}

func (f *YouPlusFile) Sync() error {
	return os.ErrPermission
}

func (f *YouPlusFile) Truncate(size int64) error {
	return os.ErrPermission
}

func (f *YouPlusFile) WriteString(s string) (ret int, err error) {
	return 0, os.ErrPermission
}
