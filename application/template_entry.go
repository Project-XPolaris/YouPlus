package application

import "github.com/projectxpolaris/youplus/service"

type EntryTemplate struct {
	Name   string      `json:"name"`
	State  string      `json:"state"`
	Export interface{} `json:"export"`
}

func (t *EntryTemplate) Assign(entry *service.Entry) {
	t.Name = entry.Name
	t.Export = entry.Export
	t.State = entry.Status
}
