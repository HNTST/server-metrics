package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

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

func getDockerContainersInfo(cli *client.Client) []ContainerInfo {
	var res []ContainerInfo
	// В новых версиях API (1.44+) используем ListOptions из подпакета container
	options := container.ListOptions{All: true}

	// ВАЖНО: В Go 1.25+ и новом SDK возвращается структура ContainerListResult
	cont, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		log.Println("Ошибка получения списка контенеров:", err)
	}
	for _, c := range cont.Containers {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		res = append(res, ContainerInfo{
			ID:    c.ID[:12],
			Name:  name,
			State: c.State,
			Image: c.Image,
		})
	}
	return res
}

func main() {
	var (
		serverURL, hostName string
		interval            int
	)

	log.Println("Инициализация агента")

	log.Println("Загрузка .env")
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден, использую дефолтные настройки")
	}

	log.Println("Чтение SERVER_URL")
	serverURL = os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080/metrics"
		log.Println("URL не найден, взят: ", serverURL)
	}

	log.Println("Чтение HOST_NAME")
	hostName = os.Getenv("HOST_NAME")
	if hostName == "" {
		h, _ := host.Info()
		hostName = h.Hostname
		log.Println("HOST_NAME не найден, взят: ", hostName)
	}

	intervalStr := os.Getenv("INTERVAL")
	interval, err = strconv.Atoi(intervalStr)
	if err != nil {
		interval = 10
	}

	scanDockerStr := os.Getenv("SCAN_DOCKER")
	cl, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Println("Не удалось запустить Docker SDK:", err)
	}

	log.Println("-------------------------")
	log.Println("SERVER_URL = ", serverURL)
	log.Println("HOST_NAME = ", hostName)
	log.Println("Интервал в секундах: ", interval)
	log.Println("Сканирование докер: ", scanDockerStr)
	log.Println("-------------------------")
	log.Println("Запуск")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var conts []ContainerInfo
		if scanDockerStr == "true" {
			conts = getDockerContainersInfo(cl)
		}

		percentages, _ := cpu.Percent(time.Second, false)
		mm, _ := mem.VirtualMemory()
		dk, _ := disk.Usage("/")
		post := Metrics{
			Host:       hostName,
			Cpu:        percentages[0],
			MemTotal:   mm.Total,
			MemUsage:   mm.Used,
			SpaceTotal: dk.Total,
			SpaceUsage: dk.Used,
			Containers: conts,
			SentAt:     time.Now(),
		}
		js, _ := json.Marshal(post)
		resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(js))
		if err != nil {
			log.Println("Ошибка отправки:", err)
			continue
		}
		resp.Body.Close()
	}
}
