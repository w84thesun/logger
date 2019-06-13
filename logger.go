package logger

import "fmt"

type Logger struct {
}

func (l *Logger) Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
