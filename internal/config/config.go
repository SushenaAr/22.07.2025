package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	Port           string
	AdvertisedAddr string
	Protocol       string
}

var (
	cfg  *Config
	once sync.Once
)

const (
	MaxTasks       int = 3
	MaxFilesInTask     = 3
)

// LoadConfig Singleton
func LoadConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil { //pushing .env into os.Getenv
			log.Println(".env not founded, continue with default values")
		}

		cfg = &Config{
			AppEnv:         getEnv("APP_ENV", "development"),
			Port:           getEnv("PORT", "8080"),
			AdvertisedAddr: getEnv("AD_ADDR", "127.0.0.1:8080"),
			Protocol:       getEnv("PROTOCOL", "http"),
		}
	})
	return cfg
}

// getEnv returning env or default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
