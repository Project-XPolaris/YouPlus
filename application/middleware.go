package application

import (
	"errors"
	"strings"

	"github.com/allentom/haruka"
)

var noAuthExactPaths = []string{
	"/user/auth",
	"/admin/auth",
	"/app/icon",
	"/notification",
	"/info",
	"/entry",
	"/dav",
}

var noAuthPrefixes = []string{
	"/dashboard/",
}

type AuthMiddleware struct {
}

func (m *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	for _, targetPath := range noAuthExactPaths {
		if ctx.Pattern == targetPath {
			return
		}
	}
	for _, prefix := range noAuthPrefixes {
		if strings.HasPrefix(ctx.Pattern, prefix) {
			return
		}
	}
	if _, hasAuth := ctx.Param["claims"]; !hasAuth {
		ctx.Abort()
		AbortErrorWithStatus(errors.New("need auth"), ctx, 403)
	}
	//service.ParseUser()
}
