package application

import (
	"errors"
	"github.com/allentom/haruka"
)

var noAuthPath = []string{
	"/user/auth",
	"/admin/auth",
	"/app/icon",
	"/notification",
}

type AuthMiddleware struct {
}

func (m *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	for _, targetPath := range noAuthPath {
		if ctx.Request.URL.Path == targetPath {
			return
		}
	}
	if _, hasAuth := ctx.Param["claims"]; !hasAuth {
		ctx.Abort()
		ctx.Interrupt()
		AbortErrorWithStatus(errors.New("need auth"), ctx, 403)
	}
}
