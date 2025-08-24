package application

import (
	"strings"

	"github.com/allentom/haruka"
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youplus/service"
)

type SystemServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

var getSystemServicesStatusHandler haruka.RequestHandler = func(ctx *haruka.Context) {
	namesStr := ctx.GetQueryString("names")
	if strings.TrimSpace(namesStr) == "" {
		ctx.JSON(haruka.JSON{
			"success":  true,
			"services": []SystemServiceStatus{},
		})
		return
	}
	names := make([]string, 0)
	for _, n := range strings.Split(namesStr, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			names = append(names, n)
		}
	}
	result := make([]SystemServiceStatus, 0, len(names))
	for _, name := range names {
		s, err := service.GetServiceByName(name)
		statusText := "Unknown"
		if err == nil && s != nil {
			if st, e := s.Status(); e == nil {
				switch st {
				case srv.StatusRunning:
					statusText = "Running"
				case srv.StatusStopped:
					statusText = "Stopped"
				default:
					statusText = "Unknown"
				}
			}
		}
		result = append(result, SystemServiceStatus{Name: name, Status: statusText})
	}
	ctx.JSON(haruka.JSON{
		"success":  true,
		"services": result,
	})
}
