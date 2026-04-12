package main

import (
	md "agent/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func PostMetrics(cfg *Config, metrics md.Metrics) error {
	js, _ := json.Marshal(metrics)
	req, err := http.NewRequest("POST", cfg.ServerURL, bytes.NewBuffer(js))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp != nil {
		log.Printf("Ответ сервера: %d", resp.StatusCode)
		resp.Body.Close()
	}
	return nil
}
