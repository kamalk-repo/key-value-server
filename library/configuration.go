package library

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
)

type AppConfig struct {
	Host          string
	Port          string
	User          string
	Password      string
	DatabaseName  string
	CacheSize     int
	ServerHost    string
	ServerPort    string
	IsConfigValid bool
}

func NewConfig() *AppConfig {
	var err error
	config := &AppConfig{}

	err = godotenv.Load()
	if err != nil {
		config.IsConfigValid = false
	} else {
		config.IsConfigValid = true
		config.Host = os.Getenv("DB_HOST")
		config.Port = os.Getenv("DB_PORT")
		config.User = os.Getenv("DB_USER")
		config.Password = os.Getenv("DB_PASSWORD")
		config.DatabaseName = os.Getenv("DB_NAME")
		config.CacheSize, err = strconv.Atoi(os.Getenv("CACHE_SIZE"))
		config.ServerHost = os.Getenv("SERVER_HOST")
		config.ServerPort = os.Getenv("SERVER_PORT")

		if err != nil || config.CacheSize <= 0 {
			config.CacheSize = 10
		}
	}

	return config
}
