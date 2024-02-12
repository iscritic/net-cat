package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Logger struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	ChatLog  *log.Logger
}

func NewLogger() *Logger {
	logFileName := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02T15-04-05"))
	logFilePath := "./logs/" + logFileName

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	infoMultiWriter := io.MultiWriter(file, os.Stdout)
	errorMultiWriter := io.MultiWriter(file, os.Stderr)
	chatWriter := io.Writer(file)

	return &Logger{
		InfoLog:  log.New(infoMultiWriter, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog: log.New(errorMultiWriter, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		ChatLog:  log.New(chatWriter, "CHAT\t", 0),
	}
}
