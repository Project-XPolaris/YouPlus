package application

import "github.com/projectxpolaris/youplus/service"

var taskTimeFormat = "2006-01-02 15:16:05"

type TaskTemplate struct {
	Id           string      `json:"id"`
	Status       string      `json:"status"`
	ErrorMessage string      `json:"errorMessage"`
	Type         string      `json:"type"`
	Updated      string      `json:"updated"`
	Created      string      `json:"created"`
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
	case *service.UnInstallAppTask:
		t.Type = "UninstallApp"
		t.Extra = task.(*service.UnInstallAppTask).Extra
	}
	t.Updated = task.GetUpdated().Format(taskTimeFormat)
	t.Created = task.GetCreated().Format(taskTimeFormat)
}

type InstallAppExtraTemplate struct {
	Output string `json:"output"`
}
