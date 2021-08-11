package application

import (
	"github.com/projectxpolaris/youplus/service"
)

type AppTemplate struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Pid       int    `json:"pid,omitempty"`
	Status    string `json:"status,omitempty"`
	AutoStart bool   `json:"autoStart"`
	Icon      string `json:"icon,omitempty"`
	Type      string `json:"type"`
}

func SerializeAppList(apps []service.App) []AppTemplate {
	data := make([]AppTemplate, 0)
	for _, app := range apps {
		meta := app.GetMeta()
		appTemplate := AppTemplate{
			Id:        meta.Id,
			Name:      meta.AppName,
			Status:    service.StatusTextMapping[meta.Status],
			AutoStart: meta.AutoStart,
			Icon:      meta.Icon,
		}
		switch app.(type) {
		case *service.ContainerApp:
			appTemplate.Type = "Container"
		case *service.ServiceApp:
			appTemplate.Type = "Service"
		case *service.RunnableApp:
			appTemplate.Type = "Runnable"
		}
		data = append(data, appTemplate)
	}
	return data
}
