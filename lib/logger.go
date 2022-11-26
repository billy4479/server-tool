package lib

import (
	"fmt"

	"github.com/fatih/color"
)

type sink interface {
	Print(...interface{}) (int, error)
	Printf(string, ...interface{}) (int, error)
	Println(...interface{}) (int, error)
}

type debugSink struct{}

func (s *debugSink) Print(a ...interface{}) (int, error) {
	return fmt.Print(a...)
}

func (s *debugSink) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}

func (s *debugSink) Println(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}

type Logger struct {
	Debug sink
	Info  sink
	Ok    sink
	Warn  sink
	Error sink
}

var L Logger = Logger{
	Debug: &debugSink{},
	Info:  color.New(color.FgBlue),
	Ok:    color.New(color.FgGreen),
	Warn:  color.New(color.FgYellow),
	Error: color.New(color.FgRed),
}
