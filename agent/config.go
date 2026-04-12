package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v3/host"
)

type Config struct {
	ServerURL  string
	HostName   string
	Interval   int
	ScanDocker bool
	APIToken   string
}

func LoadConfig(path string) *Config {
	_ = godotenv.Load(path)
	interval, _ := strconv.Atoi(getEnv("INTERVAL", "60"))
	scanDocker, _ := strconv.ParseBool(getEnv("SCAN_DOCKER", "false"))

	h, _ := host.Info()

	return &Config{
		ServerURL:  getEnv("SERVER_URL", "http://localhost:8080/metrics"),
		HostName:   getEnv("HOST_NAME", h.Hostname),
		Interval:   interval,
		ScanDocker: scanDocker,
		APIToken:   getEnv("API_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
