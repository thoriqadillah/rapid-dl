package rapid

type (
	Logger interface {
		Print(...interface{})
	}

	LoggerFunc func(setting Setting) Logger
)

var loggermap = make(map[string]LoggerFunc)

func NewLogger(provider string, setting Setting) Logger {
	logger := loggermap[provider]
	return logger(setting)
}

func RegisterLogger(name string, impl LoggerFunc) {
	loggermap[name] = impl
}
