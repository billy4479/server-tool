package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type sink struct {
	Print   func(...interface{}) (int, error)
	Printf  func(string, ...interface{}) (int, error)
	Println func(...interface{}) (int, error)
}

type Logger struct {
	Debug sink
	Info  sink
	Ok    sink
	Warn  sink
	Error sink
}

var L *Logger = nil

func SetupLogger() error {
	if L != nil {
		return nil
	}

	if C == nil {
		panic("Config has to be loaded first")
	}

	err := MakeCacheDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(C.Application.CacheDir, "logs"), 0700)
	if err != nil {
		return err
	}

	logFile, err := os.Create(filepath.Join(C.Application.CacheDir, "logs", time.Now().Format(time.RFC3339)+".log"))
	if err != nil {
		return err
	}

	makeSink := func(c *color.Color) sink {
		if c == nil {
			writer := io.MultiWriter(logFile, os.Stdout)
			return sink{
				Print: func(i ...interface{}) (int, error) {
					return fmt.Fprint(writer, i...)
				},
				Printf: func(s string, i ...interface{}) (int, error) {
					return fmt.Fprintf(writer, s, i...)
				},
				Println: func(i ...interface{}) (int, error) {
					return fmt.Fprintln(writer, i...)
				},
			}
		} else {
			return sink{
				Print: func(i ...interface{}) (int, error) {
					c.Print(i...)
					return fmt.Fprint(logFile, i...)
				},
				Printf: func(s string, i ...interface{}) (int, error) {
					c.Printf(s, i...)
					return fmt.Fprintf(logFile, s, i...)
				},
				Println: func(i ...interface{}) (int, error) {
					c.Println(i...)
					return fmt.Fprintln(logFile, i...)
				},
			}
		}
	}

	L = &Logger{
		Debug: makeSink(nil),
		Info:  makeSink(color.New(color.FgBlue)),
		Ok:    makeSink(color.New(color.FgGreen)),
		Warn:  makeSink(color.New(color.FgYellow)),
		Error: makeSink(color.New(color.FgRed)),
	}
	return nil
}
