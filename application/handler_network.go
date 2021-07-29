package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
)

var networkStatusHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"success":  true,
		"networks": service.DefaultNetworkManager.Interfaces,
	})
}
