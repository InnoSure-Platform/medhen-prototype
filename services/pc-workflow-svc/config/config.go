package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	ServiceName string      `mapstructure:"service_name"`
	Port        int         `mapstructure:"port"`
	GrpcPort    int         `mapstructure:"grpc_port"`
	Database    DatabaseCfg `mapstructure:"database"`
	Kafka       KafkaCfg    `mapstructure:"kafka"`
	Temporal    TemporalCfg `mapstructure:"temporal"`
}

type DatabaseCfg struct {
	URL          string `mapstructure:"url"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type KafkaCfg struct {
	Brokers       []string `mapstructure:"brokers"`
	SchemaRegistry string   `mapstructure:"schema_registry"`
	TopicPrefix   string   `mapstructure:"topic_prefix"`
}

type TemporalCfg struct {
	Address   string `mapstructure:"address"`
	Namespace string `mapstructure:"namespace"`
	TaskQueue string `mapstructure:"task_queue"`
}

// Load reads the configuration from the environment and default values.
func Load() (*Config, error) {
	viper.SetDefault("service_name", "pc-workflow-svc")
	viper.SetDefault("port", 8080)
	viper.SetDefault("grpc_port", 9090)
	
	viper.SetDefault("database.max_open_conns", 50)
	viper.SetDefault("database.max_idle_conns", 10)
	
	viper.SetDefault("temporal.namespace", "medhen-workflow")
	viper.SetDefault("temporal.task_queue", "workflow-task-queue")

	viper.SetEnvPrefix("WFSVC")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
