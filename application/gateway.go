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
	referer := request.Header.Get("Referer")
	if len(referer) == 0 {
		referer = request.RequestURI
	}
	targetUrl, _ := url.Parse(referer)
	referer = targetUrl.Path
	refererParts := strings.Split(referer, "/")[1:]
	entityName := refererParts[0]
	entity := service.DefaultRegisterManager.GetOnlineEntryByName(entityName)
	if entity == nil {
		AbortErrorWithStatusInWriter(errors.New("entity not found"), writer, http.StatusBadGateway)
		return
	}
	if entity.Export.Urls == nil || len(entity.Export.Urls) == 0 {
		AbortErrorWithStatusInWriter(errors.New("entity cannot access"), writer, http.StatusBadGateway)
		return
	}
	entityPath := "/"
	parts := strings.Split(request.URL.Path, "/")[1:]
	// remove prefix url of
	if len(parts) > 1 && parts[0] == entityName {
		parts = parts[1:]
	}
	if len(parts) > 0 {
		entityPath = strings.Join(parts, "/")
	}
	remoteUrl := entity.Export.Urls[0]
	remote, err := url.Parse(remoteUrl)
	if err != nil {
		panic(err)
	}
	request.URL.Path = entityPath
	request.RequestURI = entityPath
	request.Host = remote.Host
	httputil.NewSingleHostReverseProxy(remote).ServeHTTP(writer, request)
}
