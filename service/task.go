package service

import (
	"github.com/rs/xid"
	"sync"
	"time"
)

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
	GetCreated() time.Time
	GetUpdated() time.Time
}
type BaseTask struct {
	Id           string
	Status       string
	ErrorMessage string
	Created      time.Time
	Updated      time.Time
}

func (t *BaseTask) GetCreated() time.Time {
	return t.Created
}

func (t *BaseTask) GetUpdated() time.Time {
	return t.Updated
}

func (t *BaseTask) GetId() string {
	return t.Id
}

func (t *BaseTask) GetStatus() string {
	return t.Status
}

func (t *BaseTask) GetErrorMessage() string {
	return t.ErrorMessage
}
func (t *BaseTask) SetError(err error) {
	t.ErrorMessage = err.Error()
	t.Status = TaskStatusError
	t.Updated = time.Now()
}
func (t *BaseTask) SetStatus(status string) {
	t.Status = status
	t.Updated = time.Now()
}
func NewBaseTask() BaseTask {
	id := xid.New().String()
	return BaseTask{
		Id:      id,
		Status:  TaskStatusRunning,
		Created: time.Now(),
		Updated: time.Now(),
	}
}

type TaskPool struct {
	Tasks []Task
	sync.Mutex
}
