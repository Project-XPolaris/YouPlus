package service

type PathItem struct {
	Name     string `json:"name"`
	RealPath string `json:"realPath"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}
