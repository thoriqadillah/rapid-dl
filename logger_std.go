package rapid

import "log"

type stdLogger struct{}

var stdLog = "std"

// StdLogger will log into std out
func newStdLogger(_ Setting) Logger {
	return &stdLogger{}
}

func (l *stdLogger) Print(args ...interface{}) {
	log.Println(args...)
}

func init() {
	RegisterLogger(stdLog, newStdLogger)
}
