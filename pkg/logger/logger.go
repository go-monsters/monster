package logger

type Logger interface {
	Info(msg string, params ...interface{})
	Warn(msg string, params ...interface{})
	Error(msg string, params ...interface{})
}
