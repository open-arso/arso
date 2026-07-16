package node

import (
	"time"
)

type State string

const (
	StateUp   State = "up"
	StateDown State = "down"
)

type Memory struct {
	UsedMB  uint64  `json:"used_mb"`
	TotalMB uint64  `json:"total_mb"`
	Percent float64 `json:"percent"`
}

type Disk struct {
	UsedGB  uint64  `json:"used_gb"`
	TotalGB uint64  `json:"total_gb"`
	Percent float64 `json:"percent"`
}

type CPU struct {
	UsagePercent float64 `json:"used_percent"`
}

type Runtime struct {
	GoRoutines int    `json:"go_routines"`
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
}

type Status struct {
	State     State     `json:"state"`
	StartedAt time.Time `json:"started_at"`
	Uptime    string    `json:"uptime"`
	Memory    Memory    `json:"memory"`
	Disk      Disk      `json:"disk"`
	CPU       CPU       `json:"cpu"`
	Runtime   Runtime   `json:"runtime"`
}
