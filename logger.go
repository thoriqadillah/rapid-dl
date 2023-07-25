package rapid

import (
	"log"
	"sync"
)

type (
	Logger interface {
		Print(...interface{})
	}

	LoggerFunc func(setting Setting) Logger
)

var loggermap = make(map[string]LoggerFunc)
var instance sync.Map

func NewLogger(setting Setting) Logger {
	val, ok := instance.Load(setting.LoggerProvider())
	if ok {
		return val.(Logger)
	}

	logger, ok := loggermap[setting.LoggerProvider()]
	if !ok {
		log.Panicf("Provider %s is not implemented", setting.LoggerProvider())
		return nil
	}

	l := logger(setting)
	instance.Store(setting.LoggerProvider(), l)

	return l
}

func RegisterLogger(name string, impl LoggerFunc) {
	loggermap[name] = impl
}
