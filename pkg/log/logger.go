package log

type Logger interface {
	Info(msg ...any)
	Warning(msg ...any)
	Error(msg ...any)
}
