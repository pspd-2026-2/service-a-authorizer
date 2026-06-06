package config

import "os"

type Config struct {
	HTTPPort string
	ServiceName string
	LogLevel string
}

func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8081"),
		ServiceName: getEnv("SERVICE_NAME", "authorization-service"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}