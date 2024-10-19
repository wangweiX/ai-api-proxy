package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config config struct
type Config struct {
	RateLimit            string            `mapstructure:"rate_limit"`
	MaxRequestBodySizeMB int               `mapstructure:"max_request_body_size_mb"`
	FixedRequestIP       string            `mapstructure:"fixed_request_ip"`
	LogDir               string            `mapstructure:"log_dir"`
	LogName              string            `mapstructure:"log_name"`
	LogLevel             string            `mapstructure:"log_level"`
	PathMap              map[string]string `mapstructure:"path_map"`
}

// LoadConfig load config
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)

	// Set default values
	viper.SetDefault("rate_limit", "100-M")
	viper.SetDefault("max_request_body_size_mb", "100")
	viper.SetDefault("log_dir", "logs")
	viper.SetDefault("log_name", "app.log")
	viper.SetDefault("log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file failed: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %v", err)
	}

	if len(cfg.PathMap) == 0 {
		return nil, fmt.Errorf("path map config is empty")
	}

	return &cfg, nil
}
