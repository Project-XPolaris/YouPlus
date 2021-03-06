package database

import "gorm.io/gorm"

type ShareFolder struct {
	gorm.Model
	Name          string
	ZFSStorageId  string
	ZFSStorage    *ZFSStorage
	PartStorageId string
	PartStorage   *ZFSStorage
}

func GetShareFolderByName(name string) (*ShareFolder, error) {
	folder := &ShareFolder{}
	err := Instance.Model(&ShareFolder{}).Where("name = ?", name).Find(folder).Error
	return folder, err
}
func CountShareFolderByName(name string) (int64, error) {
	var count int64
	err := Instance.Model(&ShareFolder{}).Where("name = ?", name).Count(&count).Error
	return count, err
}
