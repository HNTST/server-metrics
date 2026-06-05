package main

import (
	md "agent/models"
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func formatLogsMessage(m md.Metrics) string {
	var conts string
	for _, cont := range m.Containers {
		conts += fmt.Sprintf("• %s (%s): %s\n", cont.Name, cont.Image, cont.State)
	}
	memUsedGB := float64(m.MemUsage) / (1024 * 1024 * 1024)
	memTotalGB := float64(m.MemTotal) / (1024 * 1024 * 1024)
	diskUsedGB := float64(m.SpaceUsage) / (1024 * 1024 * 1024)
	diskTotalGB := float64(m.SpaceTotal) / (1024 * 1024 * 1024)

	finalText := fmt.Sprintf(
		"🌐 Хост: %s\n"+
			"🆔 ID Агента: %d\n"+
			"📊 CPU: %.2f%%\n"+
			"🧠 RAM: %.2f / %.2f GB\n"+
			"💾 Диск: %.2f / %.2f GB\n\n"+
			"🐳 Контейнеры:\n%s\n"+
			"🕒 Время: %s",
		m.Host,
		m.AgentID,
		m.Cpu,
		memUsedGB, memTotalGB,
		diskUsedGB, diskTotalGB,
		conts,
		m.SentAt.Format("15:04:05 02.01.2006"),
	)
	return finalText
}

func Handler_Hello(bot *telego.Bot, chatID telego.ChatID) {
	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:      chatID,
		Text:        "Добро пожаловать! Я бот для просмотра метрик вашего сервера. Выберите нужный пункт в меню",
		ReplyMarkup: GetMainMenu(),
	})
}
func Handler_GetMetrics(bot *telego.Bot, id telego.ChatID, severURL string) {

	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: id,
		Text:   "⏳ Запрашиваю данные с сервера...",
	})

	metrics, err := GetMetrics(severURL)
	if err != nil {
		bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: id,
			Text:   "❌ Ошибка при получении метрик: " + err.Error(),
		})
		return
	}

	var finalText string
	finalText += "✅ Метрики получены:\n\n"
	for _, val := range metrics {
		text := formatLogsMessage(val)
		finalText += text + "\n\n" + "────────────────────" + "\n\n"
	}

	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:      id,
		Text:        finalText,
		ReplyMarkup: GetMainMenu(),
	})
}
func Handler_GetAgentList(bot *telego.Bot, id telego.ChatID, severURL string) {
	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: id,
		Text:   "⏳ Запрашиваю данные с сервера...",
	})
	agents, err := GetAgents(severURL)
	if err != nil {
		bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: id,
			Text:   "❌ Ошибка при получении метрик: " + err.Error(),
		})
		return
	}
	var agentsSTR string
	for _, agent := range agents {
		// 1. Форматируем время: будет 15:04 11.04.2026
		lastSeenFmt := agent.LastSeen.Format("15:04 02.01.2006")

		// 2. Собираем строку для одного агента
		// Используем %d для ID (так как это int)
		agentsSTR += fmt.Sprintf(
			"🆔 ID: %d\n"+
				"🖥 Имя: %s\n"+
				"🌐 IP: %s\n"+
				"🕒 В сети: %s\n"+
				"---------------------------\n",
			agent.ID,
			agent.AgentName,
			agent.AgentIP,
			lastSeenFmt,
		)
	}

	// 3. Формируем финальное сообщение
	// Если агентов нет, можно добавить проверку
	if len(agents) == 0 {
		agentsSTR = "Агенты не найдены 🤷‍♂️"
	}

	final := fmt.Sprintf("📋 Список доступных агентов:\n\n%s", agentsSTR)

	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: id,
		Text:   final,
	})
}
func Handler_GetDaily(bot *telego.Bot, id telego.ChatID, severURL string) {
	var finalText string
	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: id,
		Text:   "⏳ Запрашиваю данные с сервера...",
	})
	daily_arr, err := GetDaily(severURL)
	if err != nil {
		bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: id,
			Text:   "❌ Ошибка при получении отчета: " + err.Error(),
		})
		return
	}
	finalText += "✅ Отчеты получены:\n\n"
	for _, daily := range daily_arr {

		avg_mem := float64(daily.AvgMEM) / (1024 * 1024 * 1024)
		// avg_disk := float64(daily.AvgSpace) / (1024 * 1024 * 1024)
		max_mem := float64(daily.MaxMEM) / (1024 * 1024 * 1024)

		text := fmt.Sprintf(
			"🆔 ID Агента: %d\n"+
				"📊 AVG CPU: %.2f%%\n"+
				"📊 MAX CPU: %.2f%%\n"+
				"🧠 AVG RAM: %.2f GB\n"+ // Убрал лишний плейсхолдер для простоты
				"🧠 MAX RAM: %.2f GB\n"+
				"🕒 Дата отчета: %s",
			daily.AgentID,                         // 1
			daily.AvgCPU,                          // 2
			daily.MaxCPU,                          // 3
			avg_mem,                               // 4
			max_mem,                               // 5
			daily.ReportDate.Format("02.01.2006"), // 6 (для даты время обычно не нужно)
		)
		finalText += text + "\n\n" + "────────────────────" + "\n\n"
	}

	bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: id,
		Text:   finalText,
	})
}
