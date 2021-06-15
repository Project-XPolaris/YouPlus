package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var ReadDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	items, err := service.DefaultAddressConverterManager.ReadDir(target)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	context.JSON(items)
}
var GetRealPathHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	realPath, err := service.DefaultAddressConverterManager.GetRealPath(target)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"path": realPath,
	})
}
