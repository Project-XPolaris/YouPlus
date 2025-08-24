package application

import (
	"net/http"
	"strconv"
	"time"

	"errors"

	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/zcalusic/sysinfo"
)

var getSystemInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	context.JSON(si)
}

// GET /system/hardware
var getHardwareInfo haruka.RequestHandler = func(context *haruka.Context) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	context.JSON(haruka.JSON{
		"success":  true,
		"hardware": si,
	})
}

var getSystemMonitor haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"success": true,
		"monitor": service.DefaultMonitor.Monitor,
	})
}

// GET /system/processes?limit=50&sample=1000&search=xxx
var getSystemProcesses haruka.RequestHandler = func(context *haruka.Context) {
	limit := 50
	if l := context.GetQueryString("limit"); len(l) > 0 {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	sampleMs := 800
	if s := context.GetQueryString("sample"); len(s) > 0 {
		if v, err := strconv.Atoi(s); err == nil && v >= 100 {
			sampleMs = v
		}
	}
	search := context.GetQueryString("search")
	processes, err := service.GetProcessSnapshot(context.Request.Context(), time.Duration(sampleMs)*time.Millisecond, limit, search)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"list":    processes,
	})
}

// GET /system/sensors
var getSystemSensors haruka.RequestHandler = func(context *haruka.Context) {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	type Sensor struct {
		SensorKey   string  `json:"sensorKey"`
		Temperature float64 `json:"temperature"`
		High        float64 `json:"high"`
		Critical    float64 `json:"critical"`
	}
	list := make([]Sensor, 0, len(temps))
	for _, t := range temps {
		list = append(list, Sensor{SensorKey: t.SensorKey, Temperature: t.Temperature, High: t.High, Critical: t.Critical})
	}
	context.JSON(haruka.JSON{
		"success": true,
		"sensors": list,
	})
}

// GET /system/load
var getSystemLoad haruka.RequestHandler = func(context *haruka.Context) {
	avg, err := load.Avg()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"load1":   avg.Load1,
		"load5":   avg.Load5,
		"load15":  avg.Load15,
	})
}

// GET /system/uptime
var getSystemUptime haruka.RequestHandler = func(context *haruka.Context) {
	up, err := host.Uptime()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":   true,
		"uptimeSec": up,
	})
}

// GET /system/filesystems
var getSystemFilesystems haruka.RequestHandler = func(context *haruka.Context) {
	parts, err := disk.Partitions(false)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	type FS struct {
		Device  string  `json:"device"`
		Mount   string  `json:"mount"`
		Fstype  string  `json:"fstype"`
		Total   uint64  `json:"total"`
		Used    uint64  `json:"used"`
		Free    uint64  `json:"free"`
		UsedPct float64 `json:"usedPercent"`
	}
	list := make([]FS, 0)
	for _, p := range parts {
		if p.Mountpoint == "" {
			continue
		}
		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		list = append(list, FS{
			Device:  p.Device,
			Mount:   p.Mountpoint,
			Fstype:  p.Fstype,
			Total:   u.Total,
			Used:    u.Used,
			Free:    u.Free,
			UsedPct: u.UsedPercent,
		})
	}
	context.JSON(haruka.JSON{
		"success":     true,
		"filesystems": list,
	})
}

// GET /system/netio
var getSystemNetIO haruka.RequestHandler = func(context *haruka.Context) {
	ios, err := gnet.IOCounters(true)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	type IFace struct {
		Name        string `json:"name"`
		BytesSent   uint64 `json:"bytesSent"`
		BytesRecv   uint64 `json:"bytesRecv"`
		PacketsSent uint64 `json:"packetsSent"`
		PacketsRecv uint64 `json:"packetsRecv"`
		Errin       uint64 `json:"errsIn"`
		Errout      uint64 `json:"errsOut"`
	}
	list := make([]IFace, 0, len(ios))
	for _, io := range ios {
		list = append(list, IFace{
			Name:        io.Name,
			BytesSent:   io.BytesSent,
			BytesRecv:   io.BytesRecv,
			PacketsSent: io.PacketsSent,
			PacketsRecv: io.PacketsRecv,
			Errin:       io.Errin,
			Errout:      io.Errout,
		})
	}
	context.JSON(haruka.JSON{
		"success": true,
		"netio":   list,
	})
}

// GET /system/users
var listSystemUsersHandler haruka.RequestHandler = func(context *haruka.Context) {
	list, err := service.ListSystemUsersWithYouPlusFlag()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"users":   list,
	})
}

// POST /system/users/enable { username }

type EnableSystemUserBody struct {
	Username string `json:"username"`
}

var enableSystemUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body EnableSystemUserBody
	if err := context.ParseJson(&body); err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if len(body.Username) == 0 {
		AbortErrorWithStatus(errors.New("username required"), context, http.StatusBadRequest)
		return
	}
	if err := service.EnsureYouPlusUser(body.Username); err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
