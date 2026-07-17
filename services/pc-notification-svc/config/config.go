package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv          string `mapstructure:"APP_ENV"`
	GRPCPort        string `mapstructure:"GRPC_PORT"`
	HTTPPort        string `mapstructure:"HTTP_PORT"`
	PGDSN           string `mapstructure:"PG_DSN"`
	RedisURL        string `mapstructure:"REDIS_URL"`
	KafkaBrokers    string `mapstructure:"KAFKA_BROKERS"`
	TemporalAddress string `mapstructure:"TEMPORAL_ADDRESS"`
	OtelEndpoint    string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

func LoadConfig() *Config {
	viper.SetDefault("APP_ENV", "local")
	viper.SetDefault("GRPC_PORT", "50053")
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("PG_DSN", "postgres://medhen:medhen@localhost:5432/medhen")
	viper.SetDefault("REDIS_URL", "localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", "localhost:19092")
	viper.SetDefault("TEMPORAL_ADDRESS", "localhost:7233")
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return &cfg
}
