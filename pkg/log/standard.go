package log

import (
	"log"
	"os"
)

type standardLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

func NewStandard() Logger {
	return &standardLogger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime),
		errorLogger:   log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *standardLogger) Info(msg ...any) {
	l.infoLogger.Println(msg)
}

func (l *standardLogger) Warning(msg ...any) {
	l.warningLogger.Println(msg)
}

func (l *standardLogger) Error(msg ...any) {
	l.errorLogger.Println(msg)
}
