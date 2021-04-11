package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

type CreateZFSPoolRequestBody struct {
	Name  string   `json:"name"`
	Disks []string `json:"disks"`
}

var createZFSPoolHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateZFSPoolRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultZFSManager.CreatePool(body.Name, body.Disks...)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getZFSPoolListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := make([]*ZFSPoolTemplate, 0)
	for _, pool := range service.DefaultZFSManager.Pools {
		template := &ZFSPoolTemplate{}
		template.Assign(pool)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"pools": data,
	})
}

var removePoolHandler haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetQueryString("name")
	fmt.Println(name)
	err := service.DefaultZFSManager.RemovePool(name)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
