package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var shutdownHandler haruka.RequestHandler = func(context *haruka.Context) {
	err := service.Shutdown()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var rebootHandler haruka.RequestHandler = func(context *haruka.Context) {
	err := service.Reboot()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
