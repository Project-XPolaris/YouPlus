package service

import (
	"errors"
	"github.com/ahmetb/go-linq/v3"
	"sync"
	"time"
)

var DefaultRegisterManager *RegisterManager
var (
	MaxAlive           = 3000
	EntryNotFoundError = errors.New("target entry not found,please regis entry")
)

const (
	EntryStateOnline  = "online"
	EntryStateOffline = "offline"
)

type Entry struct {
	Name          string
	Instance      string
	Version       int64
	Status        string
	Export        EntityExport
	LastHeartbeat int64
}
type EntityExport struct {
	Urls  []string `json:"urls"`
	Extra interface{}
}
type RegisterManager struct {
	Entries []*Entry
	sync.Mutex
}

func LoadRegisterManager() error {
	DefaultRegisterManager = &RegisterManager{
		Entries: []*Entry{},
	}
	return nil
}
func (m *RegisterManager) GetEntryByInstance(instance string) *Entry {
	for _, entry := range m.Entries {
		if entry.Instance == instance {
			return entry
		}
	}
	return nil
}
func (m *RegisterManager) GetEntryByName(name string) *Entry {
	for _, entry := range m.Entries {
		if entry.Name == name {
			return entry
		}
	}
	return nil
}
func (m *RegisterManager) GetOnlineEntryByName(name string) *Entry {
	for _, entry := range m.Entries {
		if entry.Name == name && entry.Status == EntryStateOnline {
			return entry
		}
	}
	return nil
}
func (m *RegisterManager) RegisterApp(e *Entry) {
	m.Lock()
	defer m.Unlock()
	e.Status = EntryStateOffline
	m.Entries = append(m.Entries, e)
}
func (m *RegisterManager) UnregisterApp(instance string) {
	m.Lock()
	defer m.Unlock()
	linq.From(m.Entries).Where(func(i interface{}) bool {
		return i.(*Entry).Instance != instance
	}).ToSlice(&m.Entries)
}
func (m *RegisterManager) UpdateExport(instance string, export EntityExport) error {
	m.Lock()
	defer m.Unlock()
	targetEntry := m.GetEntryByInstance(instance)
	if targetEntry == nil {
		return EntryNotFoundError
	}
	targetEntry.Export = export
	return nil
}
func (m *RegisterManager) Heartbeat(instance string, stats string) error {
	m.Lock()
	defer m.Unlock()
	targetEntry := m.GetEntryByInstance(instance)
	if targetEntry == nil {
		return EntryNotFoundError
	}
	targetEntry.LastHeartbeat = time.Now().Unix()
	targetEntry.Status = stats
	// heartbeat life cycle
	go func() {
		lastTime := targetEntry.LastHeartbeat
		<-time.After(time.Duration(MaxAlive) * time.Millisecond)
		if lastTime == targetEntry.LastHeartbeat {
			targetEntry.Status = EntryStateOffline
		} else {
			targetEntry.Status = EntryStateOnline
		}
	}()
	return nil
}
