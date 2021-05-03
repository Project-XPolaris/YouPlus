package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
)

var tasksListHandler haruka.RequestHandler = func(context *haruka.Context) {
	templates := make([]TaskTemplate, 0)
	for _, task := range service.DefaultTaskPool.Tasks {
		template := TaskTemplate{}
		template.Assign(task)
		templates = append(templates, template)
	}
	context.JSON(map[string]interface{}{
		"tasks": templates,
	})
}
