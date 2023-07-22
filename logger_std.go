package rapid

import "log"

type stdLogger struct {
	setting Setting
}

var stdLog = "std"

// StdLogger will log into std out
func newStdLogger(setting Setting) Logger {
	return &stdLogger{
		setting: setting,
	}
}

func (l *stdLogger) Print(args ...interface{}) {
	log.Println(args...)
}

func init() {
	RegisterLogger(stdLog, newStdLogger)
}
