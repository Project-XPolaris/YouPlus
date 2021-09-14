package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var networkStatusHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"success":  true,
		"networks": service.DefaultNetworkManager.Interfaces,
	})
}

type UpdateNetworkConfigRequestBody struct {
	IPv4 *service.IPv4Config `json:"ipv4"`
	IPv6 *service.IPv6Config `json:"ipv6"`
}

var updateNetworkConfig haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetPathParameterAsString("name")
	var requestBody UpdateNetworkConfigRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.DefaultNetworkManager.UpdateConfig(name, requestBody.IPv4, requestBody.IPv6)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
