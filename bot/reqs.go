package main

import (
	md "agent/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func GetMetrics(url string) ([]md.Metrics, error) {
	apiToken := os.Getenv("API_TOKEN")
	var result []md.Metrics

	req, err := http.NewRequest("GET", url+"/metrics", nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("bad status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error to decode json response: ", err)
		return result, err
	}

	return result, nil
}

func GetAgents(url string) ([]md.Agent, error) {
	apiToken := os.Getenv("API_TOKEN")
	var result []md.Agent

	log.Printf("[API] Запрос списка агентов: %s/agent", url)

	req, err := http.NewRequest("GET", url+"/agent", nil)
	if err != nil {
		log.Printf("[ERROR] Не удалось создать запрос: %v", err)
		return result, err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Ошибка при выполнении запроса к серверу: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] Сервер вернул ошибку: %s (код: %d)", resp.Status, resp.StatusCode)
		return result, fmt.Errorf("bad status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] Ошибка декодирования JSON: %v", err)
		return result, err
	}

	log.Printf("[SUCCESS] Получено агентов: %d", len(result))
	return result, nil
}

func GetDaily(url string) ([]md.DailyMetrics, error) {
	apiToken := os.Getenv("API_TOKEN")
	var result []md.DailyMetrics

	log.Printf("[API] Запрос ежедневного отчета: %s/daily", url)

	req, err := http.NewRequest("GET", url+"/daily", nil)
	if err != nil {
		log.Printf("[ERROR] Не удалось создать запрос: %v", err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Ошибка при выполнении запроса к серверу: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] Сервер вернул ошибку: %s (код: %d)", resp.Status, resp.StatusCode)
		return result, fmt.Errorf("bad status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] Ошибка декодирования JSON: %v", err)
		return result, err
	}

	log.Printf("[SUCCESS] Отчеты получен")
	return result, nil
}
