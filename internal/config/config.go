package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort           string            `mapstructure:"server_port"`
	RateLimit            string            `mapstructure:"rate_limit"`
	MaxRequestBodySizeMB int               `mapstructure:"max_request_body_size_mb"`
	FixedRequestIP       string            `mapstructure:"fixed_request_ip"`
	LogDir               string            `mapstructure:"log_dir"`
	LogName              string            `mapstructure:"log_name"`
	LogLevel             string            `mapstructure:"log_level"`
	PathMap              map[string]string `mapstructure:"path_map"`
}

// LoadConfig 配置加载
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)

	// 设置默认值
	viper.SetDefault("server_port", "3001")
	viper.SetDefault("rate_limit", "100-M")
	viper.SetDefault("max_request_body_size_mb", "20")
	viper.SetDefault("log_dir", "logs")
	viper.SetDefault("log_name", "app.log")
	viper.SetDefault("log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件时发生错误: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("无法解析配置: %v", err)
	}

	if len(cfg.PathMap) == 0 {
		return nil, fmt.Errorf("路径映射配置为空")
	}

	return &cfg, nil
}
