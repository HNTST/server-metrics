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

	// Важно: для ListOptions
	md "models/models.go"

	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func getDockerContainersInfo(cli *client.Client) []md.ContainerInfo {
	var res []md.ContainerInfo

	cont, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		log.Println("Ошибка получения списка контенеров:", err)
	}
	for _, c := range cont.Items {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		res = append(res, md.ContainerInfo{
			ID:    c.ID[:12],
			Name:  name,
			State: string(c.State),
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
	err := godotenv.Load("agent/.env")
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
	apiToken := os.Getenv("API_TOKEN")

	log.Println("-------------------------")
	log.Println("SERVER_URL = ", serverURL)
	log.Println("HOST_NAME = ", hostName)
	log.Println("Интервал в секундах: ", interval)
	log.Println("Сканирование докер: ", scanDockerStr)
	if apiToken == "" {
		log.Println("API_TOKEN не найден")
	} else {
		log.Println("API_TOKEN найден")
	}
	log.Println("-------------------------")
	log.Println("Запуск")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var conts []md.ContainerInfo
		if scanDockerStr == "true" {
			conts = getDockerContainersInfo(cl)
		}

		percentages, _ := cpu.Percent(time.Second, false)
		mm, _ := mem.VirtualMemory()
		dk, _ := disk.Usage("/")
		post := md.Metrics{
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
		req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(js))
		if err != nil {
			log.Println("Ошибка создания запроса:", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+apiToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second} // Добавь таймаут, чтобы агент не завис
		resp, err := client.Do(req)

		// 1. Сначала проверяем ошибку сети/подключения
		if err != nil {
			log.Println("Ошибка отправки запроса на сервер:", err)
			continue // Идем на следующий цикл тикера
		}

		// 2. Только если ошибки нет, закрываем тело в конце
		// Используем defer внутри функции или просто Close() после обработки
		if resp != nil {
			log.Printf("Ответ сервера: %d", resp.StatusCode)
			resp.Body.Close()
		}

	}
}
