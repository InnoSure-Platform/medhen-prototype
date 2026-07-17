package config

import (
	"log"
	"os"
)

type Config struct {
	DBURL           string
	GRPCPort        string
	RESTPort        string
	KafkaBrokers    []string
}

func LoadConfig() *Config {
	// In reality this would use viper to load from env and yaml
	log.Println("Loading pc-underwriting-svc configuration...")
	return &Config{
		DBURL:        os.Getenv("DB_URL"),
		GRPCPort:     "50052",
		RESTPort:     "8082",
		KafkaBrokers: []string{"localhost:9092"},
	}
}
