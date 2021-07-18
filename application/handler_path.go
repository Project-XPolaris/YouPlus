package application

import (
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var ReadDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	claims := context.Param["claims"].(*jwt.StandardClaims)
	items, err := service.DefaultAddressConverterManager.ReadDir(target, claims.Id)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	context.JSON(items)
}
var GetRealPathHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	claims := context.Param["claims"].(*jwt.StandardClaims)
	realPath, err := service.DefaultAddressConverterManager.GetRealPath(target, claims.Id)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"path": realPath,
	})
}
