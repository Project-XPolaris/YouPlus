package service

import (
	"context"
	"sort"
	"strings"
	"time"

	gocpu "github.com/shirou/gopsutil/v3/cpu"
	goproc "github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo represents a snapshot of a running process and its resource usage.
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	Username   string  `json:"username"`
	CPUPercent float64 `json:"cpuPercent"`
	MemPercent float64 `json:"memPercent"`
	RSS        uint64  `json:"rss"`
	VMS        uint64  `json:"vms"`
	Status     string  `json:"status"`
	CreateTime int64   `json:"createTime"`
	Cmdline    string  `json:"cmdline"`
	Nice       int32   `json:"nice"`
	NumThreads int32   `json:"numThreads"`
}

// GetProcessSnapshot collects a snapshot of processes with CPU and memory usage.
// It samples CPU over the given sampleInterval to compute CPU percent similar to top/htop.
func GetProcessSnapshot(ctx context.Context, sampleInterval time.Duration, limit int, search string) ([]ProcessInfo, error) {
	// list processes
	plist, err := goproc.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	preTimes := map[int32]float64{}
	// total CPU times before
	ct0, err := gocpu.TimesWithContext(ctx, false)
	if err != nil || len(ct0) == 0 {
		return nil, err
	}
	for _, p := range plist {
		pid := p.Pid
		// Process may exit; ignore errors
		if t, err := p.TimesWithContext(ctx); err == nil {
			preTimes[pid] = t.Total()
		}
	}
	// sample interval
	select {
	case <-time.After(sampleInterval):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	ct1, err := gocpu.TimesWithContext(ctx, false)
	if err != nil || len(ct1) == 0 {
		return nil, err
	}
	deltaTotal := ct1[0].Total() - ct0[0].Total()
	if deltaTotal <= 0 {
		deltaTotal = 1 // avoid div by zero
	}
	results := make([]ProcessInfo, 0, len(plist))
	for _, p := range plist {
		pid := p.Pid
		var info ProcessInfo
		info.PID = pid
		// name
		if name, err := p.NameWithContext(ctx); err == nil {
			info.Name = name
		}
		// username
		if u, err := p.UsernameWithContext(ctx); err == nil {
			info.Username = u
		}
		// status (slice)
		if st, err := p.StatusWithContext(ctx); err == nil {
			info.Status = strings.Join(st, ",")
		}
		// cmdline
		if cmd, err := p.CmdlineWithContext(ctx); err == nil {
			info.Cmdline = cmd
		}
		// create time
		if ct, err := p.CreateTimeWithContext(ctx); err == nil {
			info.CreateTime = ct
		}
		// nice
		if n, err := p.NiceWithContext(ctx); err == nil {
			info.Nice = n
		}
		// threads
		if th, err := p.NumThreadsWithContext(ctx); err == nil {
			info.NumThreads = th
		}
		// mem
		if mem, err := p.MemoryInfoWithContext(ctx); err == nil {
			info.RSS = mem.RSS
			info.VMS = mem.VMS
		}
		if mp, err := p.MemoryPercentWithContext(ctx); err == nil {
			info.MemPercent = float64(mp)
		}
		// cpu percent
		if t1, err := p.TimesWithContext(ctx); err == nil {
			if t0, ok := preTimes[pid]; ok {
				pd := t1.Total() - t0
				if pd < 0 {
					pd = 0
				}
				info.CPUPercent = pd / deltaTotal * 100.0
			}
		}
		// filter by search if provided
		if len(search) > 0 {
			matched := false
			if info.Name != "" && containsIgnoreCase(info.Name, search) {
				matched = true
			}
			if !matched && info.Cmdline != "" && containsIgnoreCase(info.Cmdline, search) {
				matched = true
			}
			if !matched && info.Username != "" && containsIgnoreCase(info.Username, search) {
				matched = true
			}
			if !matched {
				continue
			}
		}
		results = append(results, info)
	}
	// sort by CPU desc by default
	sort.Slice(results, func(i, j int) bool { return results[i].CPUPercent > results[j].CPUPercent })
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func containsIgnoreCase(haystack string, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	h := []rune(haystack)
	n := []rune(needle)
	// simple case-insensitive search without allocations
	for i := 0; i+len(n) <= len(h); i++ {
		match := true
		for j := 0; j < len(n); j++ {
			rh := h[i+j]
			rn := n[j]
			if rh >= 'A' && rh <= 'Z' {
				rh = rh - 'A' + 'a'
			}
			if rn >= 'A' && rn <= 'Z' {
				rn = rn - 'A' + 'a'
			}
			if rh != rn {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
