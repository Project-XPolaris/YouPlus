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
	"/info",
	"/entry",
	"/dav",
}

type AuthMiddleware struct {
}

func (m *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	for _, targetPath := range noAuthPath {
		if ctx.Pattern == targetPath {
			return
		}
	}
	if _, hasAuth := ctx.Param["claims"]; !hasAuth {
		ctx.Abort()
		AbortErrorWithStatus(errors.New("need auth"), ctx, 403)
	}
	//service.ParseUser()
}
