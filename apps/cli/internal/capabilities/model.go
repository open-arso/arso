package capabilities

import (
	"time"
)

type NodeCapabilities struct {
	Tracking      Capability
	Imaging       Capability
	Weather       Capability
	Radio         Capability
	Prediction    Capability
	AI            Capability
	Storage       Capability
	Notifications Capability
}

type Capability struct {
	Status            CapabilityStatus
	LastStatusChanged time.Time
	Message           string
}

type CapabilityStatus string

const (
	CapabilityReady       CapabilityStatus = "ready"
	CapabilityDegraded    CapabilityStatus = "degraded"
	CapabilityUnavailable CapabilityStatus = "unavailable"
	CapabilityDisabled    CapabilityStatus = "disabled"
	CapabilityUnknown     CapabilityStatus = "unknown"
)
