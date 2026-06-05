package models

import "time"

type ContainerInfo struct {
	DB_ID       int    `json:"-"`
	LogID       int    `json:"-"`
	ContainerID string `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	State       string `json:"state"`
}
type Metrics struct {
	ID         int             `json:"id"`
	AgentID    int             `json:"agent_id"`
	Host       string          `json:"host"`
	Cpu        float64         `json:"cpu"`
	MemTotal   uint64          `json:"mem_total"`
	MemUsage   uint64          `json:"mem_usage"`
	SpaceTotal uint64          `json:"space_total"`
	SpaceUsage uint64          `json:"space_usage"`
	Containers []ContainerInfo `json:"containers,omitempty"`
	SentAt     time.Time       `json:"sent_at"`
}

type Agent struct {
	ID        int       `json:"id"`
	AgentName string    `json:"agent_name"`
	AgentIP   string    `json:"agent_ip"`
	LastSeen  time.Time `json:"last_seen"`
}

type DailyMetrics struct {
	ID         int       `json:"id"`
	AgentID    int       `json:"agent_id"`
	AvgCPU     float64   `json:"avg_cpu"`
	MaxCPU     float64   `json:"max_cpu"`
	AvgMEM     uint64    `json:"avg_mem"`
	MaxMEM     uint64    `json:"max_mem"`
	AvgSpace   uint64    `json:"avg_space"`
	ReportDate time.Time `json:"report_date"`
}
