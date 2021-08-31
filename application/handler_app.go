package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
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
	id, err := context.GetQueryInt("id")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultAppManager.RunApp(int64(id))
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var appListHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"apps": SerializeAppList(service.DefaultAppManager.Apps),
	})
}

var appStopHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetQueryInt("id")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultAppManager.StopApp(int64(id))
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type AutoStartRequestBody struct {
	Id int64 `json:"id"`
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
		"args":    ulist.InstallArgs,
	})
}

type InstallAppRequestBody struct {
	Args []*service.InstallArgs `json:"args"`
}

var installAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body InstallAppRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	claims := context.Param["claims"].(*jwt.StandardClaims)
	id := context.GetQueryString("id")
	var pack database.UploadInstallPack
	err = database.Instance.Where("id = ?", id).Find(&pack).Error
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
	}, body.Args, claims.Id)
	template := TaskTemplate{}
	template.Assign(task)
	context.JSON(template)
}

var uninstallAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetQueryInt("id")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTaskPool.NewUnInstallAppTask(int64(id), service.UnInstallAppCallback{
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
