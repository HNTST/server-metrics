package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	md "agent/models"

	"github.com/joho/godotenv"

	// Важно: для ListOptions

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// func getDockerContainersInfo(cli *client.Client) []md.ContainerInfo {
// 	var res []md.ContainerInfo

// 	cont, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
// 	if err != nil {
// 		log.Println("Ошибка получения списка контенеров:", err)
// 	}
// 	for _, c := range cont.Items {
// 		name := "unknown"
// 		if len(c.Names) > 0 {
// 			name = strings.TrimPrefix(c.Names[0], "/")
// 		}

//			res = append(res, md.ContainerInfo{
//				ID:    c.ID[:12],
//				Name:  name,
//				State: string(c.State),
//				Image: c.Image,
//			})
//		}
//		return res
//	}
func envLoad(path string) map[string]string {
	result := make(map[string]string)
	err := godotenv.Load(path)
	if err != nil {
		log.Println("Файл .env не найден, использую дефолтные настройки")
	}
	log.Println("Чтение SERVER_URL")
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080/metrics"
		log.Println("URL не найден, взят: ", serverURL)
	}
	result["SERVER_URL"] = serverURL

	log.Println("Чтение HOST_NAME")
	hostName := os.Getenv("HOST_NAME")
	if hostName == "" {
		h, _ := host.Info()
		hostName = h.Hostname
		log.Println("HOST_NAME не найден, взят: ", hostName)
	}
	result["HOST_NAME"] = hostName

	interval := os.Getenv("INTERVAL")
	if interval == "" {
		interval = "10"
		log.Println("INTERVAL не найден, взят: ", interval)
	}
	result["INTERVAL"] = interval

	scanDockerStr := os.Getenv("SCAN_DOCKER")
	result["SCAN_DOCKER"] = scanDockerStr

	apiToken := os.Getenv("API_TOKEN")
	result["API_TOKEN"] = apiToken

	return result
}

func main() {

	log.Println("Инициализация агента")

	log.Println("Загрузка .env")
	env := envLoad("agent/.env")

	server_url := env["SERVER_URL"]
	host_name := env["HOST_NAME"]
	intervalStr := env["INTERVAL"]
	interval, _ := strconv.Atoi(intervalStr)
	scan_docker := env["SCAN_DOCKER"]
	api_token := env["API_TOKEN"]
	// interval конвертировать в int

	log.Println("-------------------------")
	log.Println("SERVER_URL = ", server_url)
	log.Println("HOST_NAME = ", host_name)
	log.Println("Интервал в секундах: ", interval)
	log.Println("Сканирование докер: ", scan_docker)
	if api_token == "" {
		log.Println("API_TOKEN не найден")
	} else {
		log.Println("API_TOKEN найден")
	}
	log.Println("-------------------------")
	log.Println("Запуск")

	// cl, err := client.NewClientWithOpts(client.FromEnv)
	// if err != nil {
	// 	log.Println("Не удалось запустить Docker SDK:", err)
	// }

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var conts []md.ContainerInfo
		// if scan_docker == "true" {
		// 	conts = getDockerContainersInfo(cl)
		// }

		percentages, _ := cpu.Percent(time.Second, false)
		mm, _ := mem.VirtualMemory()
		dk, _ := disk.Usage("/")
		post := md.Metrics{
			Host:       host_name,
			Cpu:        percentages[0],
			MemTotal:   mm.Total,
			MemUsage:   mm.Used,
			SpaceTotal: dk.Total,
			SpaceUsage: dk.Used,
			Containers: conts,
			SentAt:     time.Now(),
		}
		js, _ := json.Marshal(post)
		req, err := http.NewRequest("POST", server_url, bytes.NewBuffer(js))
		if err != nil {
			log.Println("Ошибка создания запроса:", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+api_token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		if err != nil {
			log.Println("Ошибка отправки запроса на сервер:", err)
			continue
		}

		if resp != nil {
			log.Printf("Ответ сервера: %d", resp.StatusCode)
			resp.Body.Close()
		}

	}
}
