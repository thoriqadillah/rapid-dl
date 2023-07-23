package rapid

import "testing"

var s = DefaultSetting()

func TestLoggerInstance(t *testing.T) {
	l1 := NewLogger(stdLog, s)
	l2 := NewLogger(stdLog, s)

	if l1 != l2 {
		t.Error("Logger has different instance")
	}
}
