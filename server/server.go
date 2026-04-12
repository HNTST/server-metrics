package server

import (
	db "agent/database"
	md "agent/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func SaveLogs(repo db.DBRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		authHeader := r.Header.Get("Authorization")
		expectedToken := "Bearer " + os.Getenv("API_TOKEN")

		// 2. Сверяем
		if authHeader == "" || authHeader != expectedToken {
			log.Println("Попытка несанкционированного доступа!")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var dt md.Metrics
		err = json.NewDecoder(r.Body).Decode(&dt)
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
		err = repo.SaveLog(dt)
		if err != nil {
			log.Printf("Ошибка сохранения в базу данных: %v", err)
			return
		}
	}
}

func GetLastLog(repo db.DBRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		authHeader := r.Header.Get("Authorization")
		expectedToken := "Bearer " + os.Getenv("API_TOKEN")

		// 2. Сверяем
		if authHeader == "" || authHeader != expectedToken {
			log.Println("Попытка несанкционированного доступа!")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		metrics, err := repo.GetLastLog()

		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			log.Printf("Ошибка получения лога: %v", err)
			return
		}

		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			fmt.Println("Error to encode json response: ", err)
			return
		}
	}
}

func GetAgents(repo db.DBRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		authHeader := r.Header.Get("Authorization")
		expectedToken := "Bearer " + os.Getenv("API_TOKEN")

		// 2. Сверяем
		if authHeader == "" || authHeader != expectedToken {
			log.Println("Попытка несанкционированного доступа!")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		agents, err := repo.GetAgents()

		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			log.Printf("Ошибка получения лога: %v", err)
			return
		}

		err = json.NewEncoder(w).Encode(agents)
		if err != nil {
			fmt.Println("Error to encode json response: ", err)
			return
		}
	}
}

func GetDailyMetric(repo db.DBRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedToken := "Bearer " + os.Getenv("API_TOKEN")

		if authHeader == "" || authHeader != expectedToken {
			log.Println("Попытка несанкционированного доступа!")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		metrics, err := repo.GetDailyMetrics()

		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			log.Printf("Ошибка получения лога: %v", err)
			return
		}

		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			fmt.Println("Error to encode json response: ", err)
			return
		}

	}
}
