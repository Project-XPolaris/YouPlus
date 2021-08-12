package application

import (
	"encoding/json"
	"github.com/allentom/haruka"
	"github.com/sirupsen/logrus"
	"net/http"
)

func AbortErrorWithStatus(err error, context *haruka.Context, status int) {
	logrus.Error(err)
	context.Writer.Header().Set("Content-Type", "application/json")
	context.Writer.WriteHeader(status)
	context.JSON(map[string]interface{}{
		"success": false,
		"reason":  err.Error(),
	})
}

func AbortErrorWithStatusInWriter(err error, writer http.ResponseWriter, status int) {
	logrus.Error(err)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	data := haruka.JSON{
		"success": false,
		"reason":  err.Error(),
	}
	raw, _ := json.Marshal(data)
	writer.Write(raw)
}
