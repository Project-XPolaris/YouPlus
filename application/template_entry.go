package application

import "github.com/projectxpolaris/youplus/service"

type EntryTemplate struct {
	Name     string      `json:"name"`
	State    string      `json:"state"`
	Instance string      `json:"instance"`
	Version  int64       `json:"version"`
	Export   interface{} `json:"export"`
}

func (t *EntryTemplate) Assign(entry *service.Entry) {
	t.Name = entry.Name
	t.Export = entry.Export
	t.Version = entry.Version
	t.Instance = entry.Instance
	t.State = entry.Status
}
