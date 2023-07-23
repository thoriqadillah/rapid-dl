package rapid

import "log"

type stdLogger struct{}

const LoggerStdOut = "stdout"

// StdLogger will log into std out
func newStdLogger(_ Setting) Logger {
	return &stdLogger{}
}

func (l *stdLogger) Print(args ...interface{}) {
	log.Println(args...)
}

func init() {
	RegisterLogger(LoggerStdOut, newStdLogger)
}
