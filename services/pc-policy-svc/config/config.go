package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name     string `mapstructure:"name"`
		Env      string `mapstructure:"env"`
		HTTPPort int    `mapstructure:"http_port"`
		GRPCPort int    `mapstructure:"grpc_port"`
	} `mapstructure:"app"`

	Postgres struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"postgres"`

	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
	} `mapstructure:"kafka"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Defaults if not set
	if cfg.App.HTTPPort == 0 {
		cfg.App.HTTPPort = 8080
	}
	if cfg.App.GRPCPort == 0 {
		cfg.App.GRPCPort = 9090
	}

	return &cfg, nil
}
