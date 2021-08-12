package application

import (
	"errors"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var gatewayHandler = func(writer http.ResponseWriter, request *http.Request) {
	parts := strings.Split(request.URL.Path, "/")[1:]
	entityName := parts[0]
	entityPath := "/"
	if len(parts) > 1 {
		entityPath = strings.Join(parts[1:], "/")
	}
	entity := service.DefaultRegisterManager.GetOnlineEntryByName(entityName)
	if entity == nil {
		AbortErrorWithStatusInWriter(errors.New("entity not found"), writer, http.StatusBadGateway)
		return
	}
	if entity.Export.Urls == nil || len(entity.Export.Urls) == 0 {
		AbortErrorWithStatusInWriter(errors.New("entity cannot access"), writer, http.StatusBadGateway)
		return
	}
	remote, err := url.Parse(entity.Export.Urls[0])
	if err != nil {
		panic(err)
	}
	request.URL.Path = entityPath
	httputil.NewSingleHostReverseProxy(remote).ServeHTTP(writer, request)
}
