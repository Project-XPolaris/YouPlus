package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"github.com/spf13/afero"
	"net/http"
	"path/filepath"
)

var ReadDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	fileInfo, err := afero.ReadDir(service.DefaultFileSystem, target)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	items := make([]service.PathItem, 0)
	for _, file := range fileInfo {
		item := service.PathItem{
			Name:     file.Name(),
			RealPath: filepath.Join(target, file.Name()),
			Path:     filepath.Join(target, file.Name()),
		}
		if file.IsDir() {
			item.Type = "Directory"
		} else {
			item.Type = "File"
		}

		items = append(items, item)
	}
	context.JSON(items)
}
var GetRealPathHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	context.JSON(haruka.JSON{
		"path": target,
	})
}
