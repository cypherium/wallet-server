package cvm

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/cypherium/go-cypherium/log"
	"github.com/fatih/color"
)

const (
	ALL   = 0
	TRACE = 1
	DEBUG = 2
	INFO  = 3
	WARN  = 4
	ERROR = 5
)

type LoggerFactory struct{}

func (this *LoggerFactory) NewLogger(category string, level int, logfile string) *Logger {
	/*
		logdir := VM.GetSystemSetting("log.base")
		_, err := os.Stat(logdir)
		if os.IsNotExist(err) {
			os.MkdirAll(logdir, os.ModePerm)
		}
		path := logdir + "/" + logfile
		_, err = os.Stat(path)

		// create file if not exists
		if os.IsNotExist(err) {
			os.Create(path)
		} else {
			os.Remove(path)
			os.Create(path)
		}

		f, err1 := os.OpenFile(path, os.O_RDWR, 0666)
		if err1 != nil {
			Fatal("File does not exists or cannot be created")
		}
		w := bufio.NewWriter(f)
	*/
	//log := &Logger{category, level, w}
	log := &Logger{category, level, nil}
	return log
}

type Logger struct {
	category string
	level    int
	writer   *bufio.Writer
}

func (this *Logger) log(format string, args ...interface{}) {
	//	fmt.Fprintf(this.writer, format, args...)
	//	this.writer.Flush()
	msg := fmt.Sprintf(format, args...)
	log.Info("CVM", "", msg)
}

func (this *Logger) All(format string, args ...interface{}) {
	if this.level <= ALL {
		this.log(format, args...)
	}
}

func (this *Logger) Trace(format string, args ...interface{}) {
	if this.level <= TRACE {
		this.log(format, args...)
	}
}

func (this *Logger) Debug(format string, args ...interface{}) {
	if this.level <= DEBUG {
		this.log(format, args...)
	}
}

func (this *Logger) Info(format string, args ...interface{}) {
	if this.level <= INFO {
		this.log(format, args...)
	}
}

func (this *Logger) Warn(format string, args ...interface{}) {
	if this.level <= WARN {
		this.log(format, args...)
	}
}

func (this *Logger) Error(format string, args ...interface{}) {
	if this.level <= ERROR {
		this.log(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	color.New(color.FgHiRed).Fprintf(os.Stderr, "VM internal error: ")
	color.New(color.FgRed).Fprintf(os.Stderr, format+"\n\n ------------------------\n", args...)
	debug.PrintStack()
	panic("Fatal") //?? os.Exit(1)
}

func Bug(format string, args ...interface{}) {
	color.New(color.FgHiYellow).Fprintf(os.Stderr, "VM implmentation bug: ")
	color.New(color.FgYellow).Fprintf(os.Stderr, format+"\n\n ------------------------\n", args...)
	debug.PrintStack()
	panic("Bug") //?? os.Exit(2)
}

func Assert(expression bool, format string, args ...interface{}) {
	if !expression {
		color.New(color.FgHiYellow).Fprintf(os.Stderr, "VM runtime assertion violation: ")
		color.New(color.FgYellow).Fprintf(os.Stderr, format+"\n\n ------------------------\n", args...)
		debug.PrintStack()
		panic("Assert") //?? os.Exit(3)
	}
}
