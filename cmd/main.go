package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ContainerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"` // running, exited, etc.
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

func metricsHandle(w http.ResponseWriter, r *http.Request) {
	var dt Metrics
	err := json.NewDecoder(r.Body).Decode(&dt)
	if err != nil {
		fmt.Println("Error to encode json response: ", err)
		return
	}
	log.Printf("[%s] HOST: %-15s | CPU: %5.2f%% | RAM: %.2f/%.2f GB | SPACE: %.2f/%.2f GB",
		dt.SentAt.Format("15:04:05"),
		dt.Host,
		dt.Cpu,
		float64(dt.MemUsage)/1024/1024/1024,
		float64(dt.MemTotal)/1024/1024/1024,
		float64(dt.SpaceUsage)/1024/1024/1024,
		float64(dt.SpaceTotal)/1024/1024/1024,
	)
}

func main() {
	http.HandleFunc("POST /metrics", metricsHandle)
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Errorf("Init server error: %v", err)
		return
	}
}
