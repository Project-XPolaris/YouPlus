package application

import (
	"github.com/allentom/haruka"
	"github.com/zcalusic/sysinfo"
)

var getSystemInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	context.JSON(si)
}
