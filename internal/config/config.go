package config

import (
	"bufio"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	NexusURL    string
	NPMRegistry string
	MinAgeDays  float64
	ListenPort  string
}

func Load() *Config {
	loadEnvFile()

	minAgeDays, err := strconv.ParseFloat(getEnv("MIN_AGE_DAYS", "7"), 64)
	if err != nil {
		slog.Error("MIN_AGE_DAYS must be a number", "error", err)
		os.Exit(1)
	}

	return &Config{
		NexusURL:    getEnv("NEXUS_URL", "http://localhost:8081/repository/npm-proxy/"),
		NPMRegistry: getEnv("NPM_REGISTRY", "https://registry.npmjs.org/"),
		MinAgeDays:  minAgeDays,
		ListenPort:  getEnv("LISTEN_PORT", ":8080"),
	}
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
