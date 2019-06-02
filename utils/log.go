package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func BuildLogger(home, levelName string) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(levelName)
	if err != nil {
		level = logrus.ErrorLevel
	}
	logger.SetLevel(level)
	if err := os.MkdirAll(home, os.ModePerm); err != nil {
		fmt.Printf("Create Log Directory %s error: %v", home, err)
		os.Exit(-1)
	}
	logFilePath := filepath.Join(home, "error.log")
	output, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Open Log file (%s) err: %v", logFilePath, err)
		os.Exit(-1)
	}
	logger.Out = output
	return logger
}
