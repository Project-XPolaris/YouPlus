package application

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
)

var getDiskListHandler haruka.RequestHandler = func(context *haruka.Context) {
	disks := service.ReadDiskList()
	context.JSON(haruka.JSON{
		"disks": disks,
	})
}

var diskSmartHandler haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetQueryString("name")
	disk := service.GetDiskByName(name)
	if disk == nil {
		AbortErrorWithStatus(errors.New("disk not found"), context, 400)
		return
	}
	info, err := disk.GetSmartInfo()
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(info)
}
