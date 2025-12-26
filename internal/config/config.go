package config

import (
	"os"
)

type Config struct {
	DBUrl     string
	RedisAddr string
}

func Load() *Config {
	return &Config{
		DBUrl:     getEnv("DATABASE_URL", "postgres://payflow:payflow@postgres:5432/payflow?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR", "redis:6379"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
