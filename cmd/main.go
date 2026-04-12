package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	sv "agent/database"
	server "agent/server"
)

func main() {
	err := godotenv.Load("cmd/.env")
	if err != nil {
		log.Println("Предупреждение: .env не найден, используем системные переменные")
	}
	serv := &sv.Service{}
	err = serv.InitDB()
	if err != nil {
		log.Printf("Ошибка подлючения к базе данных: %v", err)
		return
	}

	generateDaily(serv)

	http.HandleFunc("GET /daily", server.GetDailyMetric(serv))
	http.HandleFunc("GET /agent", server.GetAgents(serv))
	http.HandleFunc("GET /metrics", server.GetLastLog(serv))
	http.HandleFunc("POST /metrics", server.SaveLogs(serv))
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Init server error: %v", err)
		return
	}
}

func generateDaily(sv *sv.Service) {

	daysStr := os.Getenv("OLD_LOGS")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		log.Printf("Ошибка чтения OLD_LOGS (%s), ставим дефолт 7: %v", daysStr, err)
		days = 7

		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				log.Println("--- Запуск ежедневного обслуживания ---")

				if err := sv.GenerateDailyReport(); err != nil {
					log.Println("Ошибка при создании отчета:", err)
					continue
				}

				if err := sv.CleanUpOldLogs(days); err != nil {
					log.Println("Ошибка при очистке старых логов:", err)
				}

				log.Println("--- Обслуживание завершено успешно ---")
			}
		}()
	}
}
