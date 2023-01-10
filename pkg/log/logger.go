package log

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger interface {
	Info(msg string)
	Error(err error)
	Panic(err error)
}

type FmtLogger struct {
	timeFormat  string
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewFmt(timeFormat string) *FmtLogger {
	infoLogger := log.New(os.Stdout, "INFO: ", 0)
	errorLogger := log.New(os.Stdout, "ERROR: ", log.Llongfile)

	return &FmtLogger{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		timeFormat:  timeFormat,
	}
}

func (l *FmtLogger) Info(msg string) {
	datetime := time.Now().Format(l.timeFormat)
	l.infoLogger.Println(fmt.Sprintf("%s %s", datetime, msg))
}

func (l *FmtLogger) Error(err error) {
	datetime := time.Now().Format(l.timeFormat)
	_ = l.errorLogger.Output(2, fmt.Sprintf("%s %s", datetime, err))
}

func (l *FmtLogger) Panic(err error) {
	datetime := time.Now().Format(l.timeFormat)
	msg := err.Error()
	stackTrace := string(stack())
	l.errorLogger.Println(fmt.Sprintf("%s %s; stacktrace: %s", datetime, msg, stackTrace))
}
