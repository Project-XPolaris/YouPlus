package database

type ZFSStorage struct {
	ID           string
	MountPoint   string
	Name         string
	ShareFolders []*ShareFolder `gorm:"foreignKey:ZFSStorageId"`
}

type PartStorage struct {
	ID           string
	MountPoint   string
	Name         string
	Source       string
	ShareFolders []*ShareFolder `gorm:"foreignKey:PartStorageId"`
}

type FolderStorage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}
