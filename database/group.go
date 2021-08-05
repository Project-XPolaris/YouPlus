package database

import "gorm.io/gorm"

type UserGroup struct {
	gorm.Model
	Gid           string
	WriteFolder   []*ShareFolder `gorm:"many2many:group_writeFolders;"`
	ValidFolder   []*ShareFolder `gorm:"many2many:group_validFolders;"`
	InvalidFolder []*ShareFolder `gorm:"many2many:group_invalidFolders;"`
	ReadFolder    []*ShareFolder `gorm:"many2many:group_readFolders;"`
}
