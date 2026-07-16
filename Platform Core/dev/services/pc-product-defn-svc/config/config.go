package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string `mapstructure:"service_name"`
	Version     string `mapstructure:"version"`
	Server      ServerConfig
	Database    DatabaseConfig
	Telemetry   TelemetryConfig
	Redis       RedisConfig
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type TelemetryConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

// Load reads configuration from file or environment variables.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/pc-product-defn-svc/")
	
	// Default values
	viper.SetDefault("service_name", "pc-product-defn-svc")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/medhen_product?sslmode=disable")
	viper.SetDefault("telemetry.endpoint", "localhost:4317")
	viper.SetDefault("redis.url", "redis://localhost:6379/0")

	// Read from environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Optionally read from config file (ignoring err if not found)
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
