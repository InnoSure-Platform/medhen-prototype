package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string
	Version     string
	Server      ServerConfig
	ClickHouse  ClickHouseConfig
	Telemetry   TelemetryConfig
}

type ServerConfig struct {
	Port int
}

type ClickHouseConfig struct {
	Addrs    []string
	Database string
	Username string
	Password string
}

type TelemetryConfig struct {
	Endpoint string
}

func Load() (*Config, error) {
	viper.SetDefault("ServiceName", "pc-reporting-svc")
	viper.SetDefault("Version", "1.0.0")
	viper.SetDefault("Server.Port", 8080)
	viper.SetDefault("ClickHouse.Addrs", []string{"localhost:9000"})
	viper.SetDefault("ClickHouse.Database", "reporting")
	viper.SetDefault("ClickHouse.Username", "default")
	viper.SetDefault("ClickHouse.Password", "")
	viper.SetDefault("Telemetry.Endpoint", "localhost:4317")

	viper.SetEnvPrefix("MDH_RPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
