package logr

import (
	"github.com/go-logr/logr"
)

type T interface {
	Logf(format string, args ...interface{})
}

func NewTestLogger(t T) logr.Logger {
	return TestLogger{T: t}
}

// TestLogger is a logr.Logger that prints through a testing.T object.
// Only error logs will have any effect.
type TestLogger struct {
	T
}

var _ logr.Logger = TestLogger{}

func (_ TestLogger) Info(_ string, _ ...interface{}) {
	// Do nothing.
}

func (_ TestLogger) Enabled() bool {
	return false
}

func (log TestLogger) Error(err error, msg string, args ...interface{}) {
	log.T.Logf("%s: %v -- %v", msg, err, args)
}

func (log TestLogger) V(v int) logr.Logger {
	return log
}

func (log TestLogger) WithName(_ string) logr.Logger {
	return log
}

func (log TestLogger) WithValues(_ ...interface{}) logr.Logger {
	return log
}
