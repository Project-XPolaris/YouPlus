package application

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/utils"
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
var getDiskInfo haruka.RequestHandler = func(context *haruka.Context) {
	device := context.GetQueryString("device")
	disk := service.GetDiskByName(device)
	if disk == nil {
		AbortErrorWithStatus(errors.New("disk not found"), context, 400)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"disk":    disk,
	})
}
var wipeDiskHandler haruka.RequestHandler = func(context *haruka.Context) {
	device := context.GetQueryString("device")
	disk := service.GetDiskByName(device)
	if disk == nil {
		AbortErrorWithStatus(errors.New("disk not found"), context, 400)
		return
	}
	err := utils.WipeDiskFS(device)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"message": "success",
	})
}

type AddPartitionRequest struct {
	Size string `json:"size"`
}

var addPartitionHandler haruka.RequestHandler = func(context *haruka.Context) {
	device := context.GetQueryString("device")
	input := AddPartitionRequest{}
	err := context.ParseJson(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	err = utils.CreateAppendDiskPartition(device, 83, input.Size)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"message": "success",
	})
}

type RemovePartitionRequest struct {
	Id int `json:"id"`
}

var removePartitionHandler haruka.RequestHandler = func(context *haruka.Context) {
	device := context.GetQueryString("device")
	input := RemovePartitionRequest{}
	err := context.ParseJson(&input)
	if err != nil {
		AbortErrorWithStatus(err, context, 400)
		return
	}
	err = utils.DeletePartition(device, input.Id)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"message": "success",
	})
}
