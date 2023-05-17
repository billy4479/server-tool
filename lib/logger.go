package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

type sink struct {
	Print   func(...interface{})
	Printf  func(string, ...interface{})
	Println func(...interface{})
}

type Logger struct {
	Debug sink
	Info  sink
	Ok    sink
	Warn  sink
	Error sink

	Writer io.Writer
	file   *os.File
	path   string
}

type LogType uint8

const (
	debug = iota
	info
	ok
	warn
	critical
)

func (l *Logger) GetCurrentLogPath() string {
	return l.path
}

func (l *Logger) Close() {
	l.file.Close()
}

var L *Logger = nil

func GetLogsPath() string {
	return filepath.Join(C.Application.CacheDir, "logs")
}

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

	path := filepath.Join(
		GetLogsPath(),
		strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")+".log",
	)
	logFile, err := os.Create(path)
	if err != nil {
		return err
	}

	writer := io.MultiWriter(logFile, os.Stdout)
	makeSink := func(c *color.Color, logType LogType) sink {
		level := ""
		switch logType {
		case debug:
			level = "-"
		case info:
			level = "+"
		case ok:
			level = "*"
		case warn:
			level = "!"
		case critical:
			level = "!!!"
		}

		getPrefix := func() string {
			t := time.Now()
			return fmt.Sprintf("[%02d:%02d:%02d][%s]: ", t.Hour(), t.Minute(), t.Second(), level)
		}

		if c == nil {
			return sink{
				Print: func(i ...interface{}) {
					fmt.Fprint(writer, getPrefix()+fmt.Sprint(i...))
				},
				Printf: func(s string, i ...interface{}) {
					fmt.Fprint(writer, getPrefix()+fmt.Sprintf(s, i...))
				},
				Println: func(i ...interface{}) {
					fmt.Fprint(writer, getPrefix()+fmt.Sprintln(i...))
				},
			}
		} else {
			return sink{
				Print: func(i ...interface{}) {
					s := getPrefix() + fmt.Sprint(i...)

					c.Print(s)
					fmt.Fprint(logFile, s)
				},
				Printf: func(format string, i ...interface{}) {
					s := getPrefix() + fmt.Sprintf(format, i...)

					c.Print(s)
					fmt.Fprint(logFile, s)
				},
				Println: func(i ...interface{}) {
					s := getPrefix() + fmt.Sprintln(i...)

					c.Print(s)
					fmt.Fprint(logFile, s)
				},
			}
		}
	}

	L = &Logger{
		Debug: makeSink(nil, debug),
		Info:  makeSink(color.New(color.FgBlue), info),
		Ok:    makeSink(color.New(color.FgGreen), ok),
		Warn:  makeSink(color.New(color.FgYellow), warn),
		Error: makeSink(color.New(color.FgRed), critical),

		file:   logFile,
		path:   path,
		Writer: writer,
	}
	return nil
}
