package config

import "os"

// Config 是服务启动需要的配置。
type Config struct {
	DatabaseURL string
	Port        string
}

// Load 读取环境变量，未设置时使用本地开发默认值。
func Load() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/vocab?sslmode=disable"),
		Port:        getenv("PORT", "8080"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
