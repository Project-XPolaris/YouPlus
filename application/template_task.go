package application

import "github.com/projectxpolaris/youplus/service"

type TaskTemplate struct {
	Id           string      `json:"id"`
	Status       string      `json:"status"`
	ErrorMessage string      `json:"errorMessage"`
	Type         string      `json:"type"`
	Extra        interface{} `json:"extra"`
}

func (t *TaskTemplate) Assign(task service.Task) {
	t.Id = task.GetId()
	t.Status = task.GetStatus()
	t.ErrorMessage = task.GetErrorMessage()
	switch task.(type) {
	case *service.InstallAppTask:
		t.Type = "InstallApp"
		t.Extra = task.(*service.InstallAppTask).Extra
	}
}

type InstallAppExtraTemplate struct {
	Output string `json:"output"`
}
