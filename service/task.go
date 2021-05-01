package service

import "sync"

var DefaultTaskPool = TaskPool{}
var (
	TaskStatusRunning = "Running"
	TaskStatusDone    = "Done"
	TaskStatusError   = "Error"
)

type Task interface {
	GetId() string
	GetStatus() string
	GetErrorMessage() string
}
type TaskPool struct {
	Tasks []Task
	sync.Mutex
}
