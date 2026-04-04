package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	sv "agent/database"
	server "agent/server"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Предупреждение: .env не найден, используем системные переменные")
	}

	serv := &sv.Service{}
	err = serv.InitDB()
	if err != nil {
		log.Printf("Ошибка подлючения к базе данных: %v", err)
		return
	}

	http.HandleFunc("POST /metrics", server.SaveLogs(serv))
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Init server error: %v", err)
		return
	}
}
