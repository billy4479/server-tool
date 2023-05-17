package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/klauspost/compress/zstd"
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
	Close  func()

	path string
}

type LogType uint8

const (
	debug = iota
	info
	ok
	warn
	critical
)

const latestName = "__latest.log"

func (l *Logger) GetCurrentLogPath() string {
	return l.path
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

	compressedLogPath := filepath.Join(
		GetLogsPath(),
		strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")+".log.zstd",
	)

	latestLogPath := filepath.Join(
		GetLogsPath(), latestName,
	)

	plainLogFile, err := os.Create(latestLogPath)
	if err != nil {
		return err
	}

	compressedLogFile, err := os.Create(compressedLogPath)
	if err != nil {
		return err
	}

	logEncoder, err := zstd.NewWriter(compressedLogFile)
	if err != nil {
		return err
	}

	writer := io.MultiWriter(plainLogFile, logEncoder, os.Stdout)
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
					fmt.Fprint(plainLogFile, s)
				},
				Printf: func(format string, i ...interface{}) {
					s := getPrefix() + fmt.Sprintf(format, i...)

					c.Print(s)
					fmt.Fprint(plainLogFile, s)
				},
				Println: func(i ...interface{}) {
					s := getPrefix() + fmt.Sprintln(i...)

					c.Print(s)
					fmt.Fprint(plainLogFile, s)
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

		Close: func() {
			plainLogFile.Close()
			logEncoder.Close()
			compressedLogFile.Close()
		},
		path:   latestLogPath,
		Writer: writer,
	}
	return nil
}
