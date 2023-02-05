package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"io/ioutil"
	"net/http"
	"strconv"
)

var fsCreateHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	file, err := service.DefaultFileSystem.Create(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	info, err := file.Stat()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	result := NewFileTemplate(info)
	context.JSON(result)
}

var fsMkdirHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	err := service.DefaultFileSystem.Mkdir(path, 0755)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}

var fsMkdirAllHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	err := service.DefaultFileSystem.MkdirAll(path, 0755)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}

var openFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	file, err := service.DefaultFileSystem.Open(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	info, err := file.Stat()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	result := NewFileTemplate(info)
	context.JSON(result)
}
var fsRemoveHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	err := service.DefaultFileSystem.Remove(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}
var fsRemoveAllHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	err := service.DefaultFileSystem.RemoveAll(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}
var fsRenameHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	source := context.GetQueryString("source")
	err := service.DefaultFileSystem.Rename(source, path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}

var fsReadByteHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	path := context.GetQueryString("path")
	rawOff := context.GetQueryString("off")
	var off int64 = 0
	off, err = strconv.ParseInt(rawOff, 10, 64)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	whence, err := context.GetQueryInt("whence")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	file, err := service.DefaultFileSystem.Open(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if off > 0 || whence > 0 {
		_, err := file.Seek(off, whence)
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusInternalServerError)
			return
		}
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	//context.Writer.Header().Set("Content-Type", "application/octet-stream")
	//context.Writer.Header().Set("Content-Length", string(len(content)))
	context.Writer.Write(content)
}

var fsWriteByteHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	path := context.GetQueryString("path")
	rawOff := context.GetQueryString("off")
	var off int64 = 0
	off, err = strconv.ParseInt(rawOff, 10, 64)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	whence, err := context.GetQueryInt("whence")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	rawFlag := context.GetQueryString("flag")
	flag, err := strconv.Atoi(rawFlag)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	file, err := service.DefaultFileSystem.OpenFile(path, flag, 0644)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if off > 0 || whence > 0 {
		_, err := file.Seek(off, whence)
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusInternalServerError)
			return
		}
	}

	contentToWrite, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}

	_, err = file.Write(contentToWrite)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})
}

var fsReadDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	count, err := context.GetQueryInt("count")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}

	path := context.GetQueryString("path")
	file, err := service.DefaultFileSystem.Open(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	infos, err := file.Readdir(count)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	result := make([]*FileInfoTemplate, len(infos))
	for i, info := range infos {
		result[i] = NewFileInfoTemplate(info)
	}
	context.JSON(result)
}
var fsTruncateHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	rawSize := context.GetQueryString("size")
	var size int64 = 0
	size, err := strconv.ParseInt(rawSize, 10, 64)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	file, err := service.DefaultFileSystem.Open(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = file.Truncate(size)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"status": "success",
	})

}
