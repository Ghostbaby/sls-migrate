package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Mode string `json:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	Charset      string `json:"charset"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// 如果 .env 文件不存在，尝试加载 test.env
		if err := godotenv.Load("test.env"); err != nil {
			// 忽略错误，使用系统环境变量
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 3306),
			Username:     getEnv("DB_USERNAME", "root"),
			Password:     getEnv("DB_PASSWORD", ""),
			Database:     getEnv("DB_DATABASE", "sls_migrate"),
			Charset:      getEnv("DB_CHARSET", "utf8mb4"),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		},
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

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
