package application

import (
	"errors"
	"github.com/allentom/haruka"
	zfs "github.com/bicomsystems/go-libzfs"
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
	err = service.DefaultZFSManager.CreateSimpleDiskPool(body.Name, body.Disks...)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type CreateZFSPoolWithConfRequestBody struct {
	Name string       `json:"name"`
	Conf service.Node `json:"conf"`
}

var createZFSPoolWithNodeHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateZFSPoolWithConfRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultZFSManager.CreatePoolWithNode(body.Name, body.Conf)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getZFSPoolListHandler haruka.RequestHandler = func(context *haruka.Context) {
	filter := service.ZFSPoolListFilter{}
	err := context.BindingInput(&filter)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	pools, err := service.DefaultZFSManager.GetPoolList(&filter)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	defer zfs.PoolCloseAll(pools)
	data := make([]*ZFSPoolTemplate, 0)
	for _, pool := range pools {
		template := &ZFSPoolTemplate{}
		template.Assign(pool)
		data = append(data, template)
	}
	context.JSON(haruka.JSON{
		"pools": data,
	})
}
var getZFSPoolHandler haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetPathParameterAsString("name")
	pool, err := service.DefaultZFSManager.GetPoolByName(name)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if pool == nil {
		AbortErrorWithStatus(errors.New("pool not found"), context, http.StatusNotFound)
		return
	}
	template := &ZFSPoolTemplate{}
	template.Assign(*pool)
	pool.Close()
	context.JSON(haruka.JSON{
		"data":    template,
		"success": "true",
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

var datasetListHandler haruka.RequestHandler = func(context *haruka.Context) {
	filter := service.DatasetQueryFilter{}
	err := context.BindingInput(&filter)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	datasets, err := service.DefaultZFSManager.GetAllDataset(filter)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	data := SerializerDatasetTemplates(datasets)
	service.DefaultZFSManager.CloseAllDataset(datasets)
	context.JSON(haruka.JSON{
		"list":    data,
		"success": true,
	})
}

type CreateDatasetRequestBody struct {
	Path string `json:"path"`
}

var createDatasetHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateDatasetRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	dataset, err := service.DefaultZFSManager.CreateDataset(body.Path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	template := DatasetTemplate{}
	template.Assign(&dataset)
	context.JSON(haruka.JSON{
		"success": true,
		"data":    template,
	})
}

var deleteDatasetHandler haruka.RequestHandler = func(context *haruka.Context) {
	datasetPath := context.GetQueryString("path")
	err := service.DefaultZFSManager.DeleteDataset(datasetPath)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

type CreateSnapshotRequestBody struct {
	Name    string `json:"name"`
	Dataset string `json:"dataset"`
}

var createSnapshotHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body CreateSnapshotRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	snapshot, err := service.DefaultZFSManager.CreateSnapshot(body.Dataset, body.Name)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	template := DatasetTemplate{}
	template.Assign(&snapshot)
	context.JSON(haruka.JSON{
		"success": true,
		"data":    template,
	})
}

var datasetSnapshotListHandler haruka.RequestHandler = func(context *haruka.Context) {
	datasetPath := context.GetQueryString("dataset")
	datasets, err := service.DefaultZFSManager.GetDatasetSnapshotList(datasetPath)
	if err != nil {
		AbortErrorWithStatus(err, context, 500)
		return
	}
	data := SerializerDatasetTemplates(datasets)
	service.DefaultZFSManager.CloseAllDataset(datasets)
	context.JSON(haruka.JSON{
		"list":    data,
		"success": true,
	})
}

type DatasetRollbackRequestBody struct {
	Dataset  string `json:"dataset"`
	Snapshot string `json:"snapshot"`
}

var datasetSnapshotRollbackHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body DatasetRollbackRequestBody
	err := context.ParseJson(&body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultZFSManager.DatasetRollback(body.Dataset, body.Snapshot)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var deleteSnapshotHandler haruka.RequestHandler = func(context *haruka.Context) {
	dataset := context.GetQueryString("dataset")
	snapshot := context.GetQueryString("snapshot")
	err := service.DefaultZFSManager.DeleteSnapshot(dataset, snapshot)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
