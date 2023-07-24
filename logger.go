package rapid

import "log"

type (
	Logger interface {
		Print(...interface{})
	}

	LoggerFunc func(setting Setting) Logger
)

var loggermap = make(map[string]LoggerFunc)

func NewLogger(provider string, setting Setting) Logger {
	logger, ok := loggermap[provider]
	if !ok {
		log.Panicf("Provider %s is not implemented", provider)
		return nil
	}

	return logger(setting)
}

func RegisterLogger(name string, impl LoggerFunc) {
	loggermap[name] = impl
}
