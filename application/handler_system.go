package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"github.com/zcalusic/sysinfo"
)

var getSystemInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	context.JSON(si)
}

var getSystemMonitor haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"success": true,
		"monitor": service.DefaultMonitor.Monitor,
	})
}
