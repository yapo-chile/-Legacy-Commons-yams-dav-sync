package interfaces

type Logger interface {
	Debug(string)
	Info(string)
	Warn(string)
	Error(string)
}
