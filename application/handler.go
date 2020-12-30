package application

import (
	"github.com/allentom/haruka"
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
