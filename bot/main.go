package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
)

func main() {
	err := godotenv.Load("bot/.env")
	if err != nil {
		log.Println("Файл .env не найден")
		return
	}
	serverURL := os.Getenv("SERVER_URL")

	botToken := os.Getenv("BOT_TOKEN")

	proxyURL, _ := url.Parse("http://127.0.0.1:2080")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	botAPI, err := telego.NewBot(botToken, telego.WithHTTPClient(client))
	if err != nil {
		log.Println(err)
		return
	}

	updates, _ := botAPI.UpdatesViaLongPolling(context.Background(), nil)
	for update := range updates {
		if update.Message != nil {
			chatID := telego.ChatID{ID: update.Message.Chat.ID}
			text := update.Message.Text
			log.Printf("Получено сообщение: %s от ID: %d", text, chatID)

			switch text {
			case "/start":
				Handler_Hello(botAPI, chatID)
			case "📊 Получить метрики":
				Handler_GetMetrics(botAPI, chatID, serverURL)
			case "Список агентов":
				Handler_GetAgentList(botAPI, chatID, serverURL)
			case "Получить отчет":
				Handler_GetDaily(botAPI, chatID, serverURL)
			default:
				botAPI.SendMessage(context.Background(), &telego.SendMessageParams{
					ChatID:      chatID,
					Text:        "Используйте кнопки меню для управления.",
					ReplyMarkup: GetMainMenu(),
				})
			}
		}
	}

}
