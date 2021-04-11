package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
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
		appTemplate := AppTemplate{
			Id:        app.Id,
			Name:      app.AppName,
			Status:    service.StatusTextMapping[app.Status],
			AutoStart: app.AutoStart,
			Icon:      app.Icon,
		}
		if app.Cmd != nil {
			appTemplate.Pid = app.Cmd.Process.Pid
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
	err = service.DefaultAppManager.AddApp(body.Path)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
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
		AbortErrorWithStatus(err, context, 500)
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
	err = service.DefaultUserManager.NewUser(body.Username, body.Password, false)
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
	err := service.DefaultAppManager.StopApp(id)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	err = service.DefaultStoragePool.RemoveStorage(id)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type UserAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var userLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body UserAuthRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	tokenStr, err := service.UserLogin(body.Username, body.Password)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"token":   tokenStr,
	})
}

var checkTokenHandler haruka.RequestHandler = func(context *haruka.Context) {
	rawToken := context.GetQueryString("token")
	user, err := service.ParseUser(rawToken)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success":  true,
		"username": user.Username,
		"uid":      user.Uid,
	})
}

var appIconHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	app := service.DefaultAppManager.GetAppByIdApp(id)
	if app == nil || len(app.Icon) == 0 {
		context.JSON(haruka.JSON{
			"success": false,
			"token":   "app not found",
		})
		return
	}
	http.ServeFile(context.Writer, context.Request, filepath.Join(app.Dir, app.Icon))
}
