package logger

import (
	"ai-api-proxy/internal/config"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
)

var Logger *logrus.Logger

// InitLogger 初始化日志
func InitLogger(config *config.Config) {
	if config.LogDir == "" {
		panic(fmt.Errorf("log dir is empty"))
	}
	if config.LogName == "" {
		panic(fmt.Errorf("log name is empty"))
	}
	if config.LogLevel == "" {
		panic(fmt.Errorf("log level is empty"))
	}
	if _, err := os.Stat(config.LogDir); err != nil {
		if err := os.MkdirAll(config.LogDir, 0755); err != nil {
			panic(fmt.Errorf("init log dir falied, err: %v", err))
		}
	}
	filepath := path.Join(config.LogDir, config.LogName)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("init log file falied, err: %v", err))
	}

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		panic(fmt.Errorf("init log level falied, err: %v", err))
	}

	fileAndStdoutWriter := io.MultiWriter(os.Stdout, file)

	Logger = logrus.New()
	Logger.SetOutput(fileAndStdoutWriter)
	Logger.SetFormatter(&logrus.TextFormatter{})
	Logger.SetLevel(level)
	Logger.SetReportCaller(true)

}
