package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
)

var deviceInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	userCount, _ := service.GetUserCount()
	ZFSCount, _ := service.DefaultZFSManager.GetPoolCount()
	shareFolderCount, _ := service.GetShareFolderCount()
	appCount := len(service.DefaultAppManager.Apps)
	storageCount := len(service.DefaultStoragePool.Storages)
	diskCount := len(service.ReadDiskList())
	context.JSON(haruka.JSON{
		"success":          true,
		"userCount":        userCount,
		"zfsCount":         ZFSCount,
		"shareFolderCount": shareFolderCount,
		"appCount":         appCount,
		"storageCount":     storageCount,
		"diskCount":        diskCount,
	})
}
