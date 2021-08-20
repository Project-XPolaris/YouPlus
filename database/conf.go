package database

import "gorm.io/gorm"

type ConfigItem struct {
	gorm.Model
	Name  string `json:"name"`
	Type  string `json:"type"`
	Key   string `json:"key"`
	AppId uint
	App   *App
}
