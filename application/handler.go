package application

import (
	"github.com/allentom/haruka"
	"net/http"
	"youplus/service"
)

var startAppHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	err := service.DefaultAppManager.RunApp(id)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(map[string]interface{}{
		"success": true,
	})
}

var appListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := make([]AppTemplate, 0)
	for _, app := range service.DefaultAppManager.Apps {
		appTemplate := AppTemplate{
			Id:        app.Id,
			Name:      app.AppName,
			Status:    service.StatusTextMapping[app.Status],
			AutoStart: app.AutoStart,
		}
		if app.Cmd != nil {
			appTemplate.Pid = app.Cmd.Process.Pid
		}
		data = append(data, appTemplate)
	}
	context.JSON(map[string]interface{}{
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
	context.JSON(map[string]interface{}{
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
	context.JSON(map[string]interface{}{
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
	context.JSON(map[string]interface{}{
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
	err = service.DefaultAppManager.AddApp(body.Path)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(map[string]interface{}{
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
		AbortErrorWithStatus(err, context, 500)
		return
	}
	err = service.DefaultAppManager.RemoveApp(body.Path)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(map[string]interface{}{
		"success": true,
	})
}

var createShareHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.NewShareFolderOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.CreateNewShareFolder(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getDiskListHandler haruka.RequestHandler = func(context *haruka.Context) {
	disks := service.ReadDiskList()
	context.JSON(haruka.JSON{
		"disks": disks,
	})
}

type CreateUserRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var createUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateUserRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.NewUser(body.Username, body.Password)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getUserList haruka.RequestHandler = func(context *haruka.Context) {
	userList, err := service.GetUserList()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"users": userList,
	})
}
