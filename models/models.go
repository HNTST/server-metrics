package models

import "time"

type ContainerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Image string `json:"image"`
}
type Metrics struct {
	Host       string          `json:"host"`
	Cpu        float64         `json:"cpu"`
	MemTotal   uint64          `json:"mem_total"`
	MemUsage   uint64          `json:"mem_usage"`
	SpaceTotal uint64          `json:"space_total"`
	SpaceUsage uint64          `json:"space_usage"`
	Containers []ContainerInfo `json:"containers,omitempty"`
	SentAt     time.Time       `json:"sent_at"`
}
