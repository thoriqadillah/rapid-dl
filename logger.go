package rapid

import "sync"

type (
	Logger interface {
		Print(...interface{})
	}

	LoggerFunc func(setting Setting) Logger
)

var loggermap = make(map[string]LoggerFunc)
var loggerInstance sync.Map

func NewLogger(provider string, setting Setting) Logger {
	logger := loggermap[provider]
	instance, ok := loggerInstance.Load(provider)
	if ok {
		return instance.(Logger)
	}

	l := logger(setting)
	loggerInstance.Store(provider, l)

	return l
}

func RegisterLogger(name string, impl LoggerFunc) {
	loggermap[name] = impl
}
