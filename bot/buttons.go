package main

import "github.com/mymmrac/telego"

func GetMainMenu() *telego.ReplyKeyboardMarkup {
	return &telego.ReplyKeyboardMarkup{
		Keyboard: [][]telego.KeyboardButton{
			{
				{Text: "📊 Получить метрики"},
				{Text: "Список агентов"},
				{Text: "Получить отчет"},
			},
			{
				{Text: "⚙️ Настройки"},
			},
		},
		ResizeKeyboard: true,
	}
}
