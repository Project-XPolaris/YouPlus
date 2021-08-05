package database

import "gorm.io/gorm"

type ShareFolder struct {
	gorm.Model
	Name          string
	ZFSStorageId  string
	ZFSStorage    *ZFSStorage
	PartStorageId string
	PartStorage   *ZFSStorage
	ValidUsers    []*User      `gorm:"many2many:user_validFolders;"`
	InvalidUsers  []*User      `gorm:"many2many:user_invalidFolders;"`
	WriteUsers    []*User      `gorm:"many2many:user_writeFolders;"`
	ReadUsers     []*User      `gorm:"many2many:user_readFolders;"`
	WriteGroups   []*UserGroup `gorm:"many2many:group_writeFolders;"`
	ValidGroups   []*UserGroup `gorm:"many2many:group_validFolders;"`
	InvalidGroups []*UserGroup `gorm:"many2many:group_invalidFolders;"`
	ReadGroups    []*UserGroup `gorm:"many2many:group_readFolders;"`
	Public        bool
	Enable        bool
	Readonly      bool
	Path          string
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
