package rabbitmq

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type LineFormatter struct {
}

func (l *LineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func buildRecorder(home, name string) *logrus.Logger {
	logger := logrus.New()
	if err := os.MkdirAll(home, os.ModePerm); err != nil {
		fmt.Printf("Create Log Directory %s error: %v", home, err)
		os.Exit(-1)
	}
	logFilePath := filepath.Join(home, name)
	output, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Open Log file (%s) err: %v", logFilePath, err)
		os.Exit(-1)
	}
	logger.Out = output
	logger.SetFormatter(&LineFormatter{})
	return logger
}
