package service

import (
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/sirupsen/logrus"
	"time"
)

var DefaultMonitor = SystemMonitor{}
var MonitorLogger = logrus.New().WithFields(logrus.Fields{
	"scope": "Monitor",
})

type Monitor struct {
	CPU    CpuInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
}
type MemoryInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
	Cache uint64 `json:"cache"`
}
type CpuInfo struct {
	Idle   uint64 `json:"idle"`
	Total  uint64 `json:"total"`
	User   uint64 `json:"user"`
	System uint64 `json:"system"`
	Iowait uint64 `json:"iowait"`
}
type SystemMonitor struct {
	Monitor Monitor
}

func (m *SystemMonitor) Run() {
	m.Monitor = Monitor{
		Memory: MemoryInfo{},
		CPU:    CpuInfo{},
	}
	go func() {
		var before *cpu.Stats
		for true {
			now, err := cpu.Get()
			if err != nil {
				MonitorLogger.Error(err)
				continue
			}
			if before == nil {
				before = now
				continue
			}
			m.Monitor.CPU = CpuInfo{
				Idle:   now.Idle - before.Idle,
				Total:  now.Total - before.Total,
				User:   now.User - before.User,
				System: now.System - before.System,
			}
			before = now
			// read ram info
			mem, err := memory.Get()
			if err != nil {
				MonitorLogger.Error(err)
				continue
			}
			m.Monitor.Memory = MemoryInfo{
				Total: mem.Total,
				Free:  mem.Free,
				Used:  mem.Used,
				Cache: mem.Cached,
			}
			<-time.After(1 * time.Second)
		}

	}()
}
