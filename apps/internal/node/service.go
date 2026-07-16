package node

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"time"
)

var startTime = time.Now()

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Status(ctx context.Context) (Status, error) {

	vm, err := mem.VirtualMemory()
	if err != nil {
		return Status{}, err
	}
	d, err := disk.Usage("/")
	if err != nil {
		return Status{}, err
	}
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return Status{}, err
	}

	status := Status{
		State:     StateUp,
		Uptime:    time.Since(startTime).String(),
		StartedAt: startTime,
		Memory: Memory{
			UsedMB:  vm.Used / 1024 / 1024,
			TotalMB: vm.Total / 1024 / 1024,
			Percent: vm.UsedPercent,
		},

		Disk: Disk{
			UsedGB:  d.Used / 1024 / 1024 / 1024,
			TotalGB: d.Total / 1024 / 1024 / 1024,
			Percent: d.UsedPercent,
		},

		CPU: CPU{
			UsagePercent: cpuPercent[0],
		},

		Runtime: Runtime{
			GoRoutines: runtime.NumGoroutine(),
			GoVersion:  runtime.Version(),
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
		},
	}

	return status, nil
}
