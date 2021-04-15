package application

import (
	"github.com/allentom/haruka"
	libzfs "github.com/bicomsystems/go-libzfs"
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
	pools, err := libzfs.PoolOpenAll()
	if err != nil {
		context.JSON(haruka.JSON{
			"pools": []string{},
		})
		return
	}
	for _, pool := range pools {
		template := &ZFSPoolTemplate{}
		template.Assign(pool)
		data = append(data, template)
	}
	libzfs.PoolCloseAll(pools)
	context.JSON(haruka.JSON{
		"pools": data,
	})
}

var removePoolHandler haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetQueryString("name")
	err := service.DefaultZFSManager.RemovePool(name)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
