package config

import (
	"os"
	"strconv"
)

// DatabaseConfig represents the MySQL database configuration.
type MySQLDatabaseConfig struct {
	Username     string
	Password     string
	Host         string
	Port         string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
}

func GetPrimaryMySQLDBConfig() MySQLDatabaseConfig {
	return MySQLDatabaseConfig{
		Username:     getEnv("DB_USERNAME", ""),
		Password:     getEnv("DB_PASSWORD", ""), 
		Host:         getEnv("DB_HOST", ""),
		Port:         getEnv("DB_PORT", ""),
		Database:     getEnv("DB_DATABASE_NAME", ""), 
		MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
