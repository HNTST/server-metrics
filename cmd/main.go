package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	md "models/models.go"

	"github.com/joho/godotenv"
)

func metricsHandle(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load("cmd/.env")
	if err != nil {
		log.Println("Файл .env не найден, использую дефолтные настройки")
	}
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
}

func main() {
	http.HandleFunc("POST /metrics", metricsHandle)
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Errorf("Init server error: %v", err)
		return
	}
}
