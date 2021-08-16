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

func SerializerEntityList(entityList []*service.Entry) []EntryTemplate {
	result := make([]EntryTemplate, 0)
	for _, entity := range entityList {
		template := EntryTemplate{}
		template.Assign(entity)
		result = append(result, template)
	}
	return result
}
