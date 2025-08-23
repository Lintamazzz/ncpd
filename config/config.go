package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 存储应用配置
type Config struct {
	NicoClientID     string
	NicoRefreshToken string
}

// Load 加载配置
func Load() *Config {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	config := &Config{
		NicoClientID:     getEnv("NICO_CLIENT_ID", ""),
		NicoRefreshToken: getEnv("NICO_REFRESH_TOKEN", ""),
	}

	// 验证必要的配置
	if config.NicoClientID == "" {
		log.Fatal("请设置环境变量 NICO_CLIENT_ID")
	}
	if config.NicoRefreshToken == "" {
		log.Fatal("请设置环境变量 NICO_REFRESH_TOKEN")
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
