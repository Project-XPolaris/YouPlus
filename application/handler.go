package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
	"path/filepath"
)

type NewStorageRequest struct {
	Source string `json:"source"`
	Type   string `json:"type"`
}

var newStorage haruka.RequestHandler = func(context *haruka.Context) {
	var body NewStorageRequest
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultStoragePool.NewStorage(body.Source, body.Type)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getStorageListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := make([]*StorageTemplate, 0)
	for _, storage := range service.DefaultStoragePool.Storages {
		template := &StorageTemplate{}
		template.Assign(storage)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"storages": data,
	})
}

var removeStorage haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	err := service.DefaultStoragePool.RemoveStorage(id)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var updateStorageHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetPathParameterAsString("id")
	option := service.StorageUpdateOption{}
	err := context.ParseJson(&option)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultStoragePool.UpdateStorage(id, option)
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var appIconHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetPathParameterAsInt("id")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	app := service.DefaultAppManager.GetAppByIdApp(int64(id))
	if app == nil || len(app.GetMeta().Icon) == 0 {
		context.JSON(haruka.JSON{
			"success": false,
			"token":   "app not found",
		})
		return
	}
	http.ServeFile(context.Writer, context.Request, filepath.Join(app.GetMeta().Dir, app.GetMeta().Icon))
}

var serviceInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"name":    "YouPlus service",
		"success": true,
	})
}
