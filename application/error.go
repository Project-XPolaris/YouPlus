package application

import (
	"github.com/allentom/haruka"
	"github.com/sirupsen/logrus"
)

func AbortErrorWithStatus(err error, context *haruka.Context, status int) {
	logrus.Error(err)
	context.Writer.WriteHeader(status)
	context.JSON(map[string]interface{}{
		"success": false,
		"reason":  err.Error(),
	})
}
