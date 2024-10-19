package logger

import (
	"ai-api-proxy/internal/config"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger init logger
func InitLogger(config *config.Config) error {
	if config.LogDir == "" {
		return fmt.Errorf("log dir is empty")
	}
	if config.LogName == "" {
		return fmt.Errorf("log name is empty")
	}
	if config.LogLevel == "" {
		return fmt.Errorf("log level is empty")
	}
	if _, err := os.Stat(config.LogDir); err != nil {
		if err := os.MkdirAll(config.LogDir, 0755); err != nil {
			return fmt.Errorf("init log dir falied, err: %v", err)
		}
	}
	filepath := path.Join(config.LogDir, config.LogName)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("init log file falied, err: %v", err)
	}

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("init log level falied, err: %v", err)
	}

	fileAndStdoutWriter := io.MultiWriter(os.Stdout, file)

	Logger = logrus.New()
	Logger.SetOutput(fileAndStdoutWriter)
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	Logger.SetLevel(level)
	Logger.SetReportCaller(true)

	return nil
}
