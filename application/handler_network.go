package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
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

}
