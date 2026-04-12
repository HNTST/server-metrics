package main

import (
	"log"
	"time"

	md "agent/models"

	"github.com/moby/moby/client"
)

func main() {

	log.Println("Инициализация агента")

	log.Println("Загрузка .env")
	cfg := LoadConfig("agent/.env")

	log.Println("-------------------------")
	log.Println("SERVER_URL = ", cfg.ServerURL)
	log.Println("HOST_NAME = ", cfg.HostName)
	log.Println("Интервал в секундах: ", cfg.Interval)
	log.Println("Сканирование докер: ", cfg.ScanDocker)
	log.Println("API_TOKEN: ", cfg.APIToken)
	log.Println("-------------------------")
	log.Println("Запуск")

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("Ошибка получения данных докер контейнера: %v", err)
	}
	for range ticker.C {
		var (
			conts []md.ContainerInfo
			err   error
		)
		if cfg.ScanDocker {
			conts, err = getDockerContainersInfo(cli)
			if err != nil {
				log.Printf("Ошибка получения данных докер контейнера: %v", err)
			}
		}

		metrics, err := getServerInfo(cfg, conts)
		if err != nil {
			log.Printf("Ошибка получения данных сервера: %v", err)
		}
		err = PostMetrics(cfg, metrics)
		if err != nil {
			log.Printf("Ошибка отправки данных на центральный сервер: %v", err)
		}
	}
}
