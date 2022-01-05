package main

import (
	"github.com/fatih/color"
)

// type errorClass struct{}

// func (*errorClass) Println(a ...interface{}) (int, error) {
// 	panic(a)
// 	return color.New(color.FgRed).Println(a...)
// }
// func (*errorClass) Printf(format string, a ...interface{}) (int, error) {
// 	panic(fmt.Sprintf(format, a...))
// 	return color.New(color.FgRed).Printf(format, a...)
// }

var (
	Ok    = color.New(color.FgGreen)
	Warn  = color.New(color.FgYellow)
	Error = color.New(color.FgRed)
	// Error = errorClass{}
	Info = color.New(color.FgBlue)
)
