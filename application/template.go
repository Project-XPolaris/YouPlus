package application

type AppTemplate struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Pid    int    `json:"pid"`
	Status string `json:"status"`
}
