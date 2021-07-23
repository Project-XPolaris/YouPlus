package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
	"github.com/rs/xid"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var startAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	err := service.DefaultAppManager.RunApp(id)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var appListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := make([]AppTemplate, 0)
	for _, app := range service.DefaultAppManager.Apps {
		meta := app.GetMeta()
		appTemplate := AppTemplate{
			Id:        meta.Id,
			Name:      meta.AppName,
			Status:    service.StatusTextMapping[meta.Status],
			AutoStart: meta.AutoStart,
			Icon:      meta.Icon,
		}
		data = append(data, appTemplate)
	}
	context.JSON(haruka.JSON{
		"apps": data,
	})
}

var appStopHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	err := service.DefaultAppManager.StopApp(id)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type AutoStartRequestBody struct {
	Id string `json:"id"`
}

var appSetAutoStart haruka.RequestHandler = func(context *haruka.Context) {
	var body AutoStartRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	err = service.DefaultAppManager.SetAutoStart(body.Id, true)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var appRemoveAutoStart haruka.RequestHandler = func(context *haruka.Context) {
	var body AutoStartRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	err = service.DefaultAppManager.SetAutoStart(body.Id, false)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type AddAppRequestBody struct {
	Path string `json:"path"`
}

var addAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body AddAppRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	//err = service.DefaultAppManager.AddApp(body.Path)
	//if err != nil {
	//	AbortErrorWithStatus(err, context, 500)
	//	return
	//}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type RemoveAppRequestBody struct {
	Path string `json:"path"`
}

var removeAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body RemoveAppRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultAppManager.RemoveApp(body.Path)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var uploadAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	err := context.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	file, _, err := context.Request.FormFile("file")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}

	defer file.Close()
	err = os.MkdirAll("./upload", os.ModePerm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	packageName := fmt.Sprintf("%s.upk", xid.New().String())
	packagePath := path.Join("./upload", packageName)
	dst, err := os.Create(packagePath)
	defer dst.Close()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	ulist, app, err := service.CheckInstallPack(packageName)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	pack, err := service.SaveInstallPack(packageName)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"id":      pack.ID,
		"name":    ulist.Name,
		"type":    ulist.InstallType,
		"appName": app.AppName,
	})
}
var installAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	var pack database.UploadInstallPack
	err := database.Instance.Where("id = ?", id).Find(&pack).Error
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	task := service.DefaultTaskPool.NewInstallAppTask(filepath.Join("./upload", pack.FileName), service.InstallAppCallback{
		OnDone: func(task *service.InstallAppTask) {
			template := TaskTemplate{}
			template.Assign(task)
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": InstallDoneEvent,
				"data":  template,
			})
		},
		OnError: func(task *service.InstallAppTask) {
			template := TaskTemplate{}
			template.Assign(task)
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": InstallErrorEvent,
				"data":  template,
			})
		},
	})
	template := TaskTemplate{}
	template.Assign(task)
	context.JSON(template)
}

var uninstallAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	task := service.DefaultTaskPool.NewUnInstallAppTask(id, service.UnInstallAppCallback{
		OnDone: func(task *service.UnInstallAppTask) {
			template := TaskTemplate{}
			template.Assign(task)
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": UninstallDoneEvent,
				"data":  template,
			})
		},
		OnError: func(task *service.UnInstallAppTask) {
			template := TaskTemplate{}
			template.Assign(task)
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": UninstallErrorEvent,
				"data":  template,
			})
		},
	})
	template := TaskTemplate{}
	template.Assign(task)
	context.JSON(template)
}
