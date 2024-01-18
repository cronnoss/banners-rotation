package logger

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type LogLevel string

const (
	Error   = "ERROR"
	Warning = "WARNING"
	Info    = "INFO"
	Debug   = "DEBUG"
)

type Logger struct {
	level   LogLevel
	writeTo io.Writer
}

func New(level string, writeTo io.Writer) *Logger {
	lev := logLevelFromString(level)
	return &Logger{lev, writeTo}
}

func logLevelFromString(level string) LogLevel {
	level = strings.ToLower(level)
	switch level {
	case "error":
		return Error
	case "warning":
		return Warning
	case "info":
		return Info
	case "debug":
		return Debug
	default:
		return Info
	}
}

func (l Logger) msg(level LogLevel, template string, a ...any) {
	var buildedString strings.Builder
	buildedString.WriteString(fmt.Sprintf("%s [%s] ", level, time.Now().UTC().Format("2006-01-02 15:04:05")))
	buildedString.WriteString(fmt.Sprintf(template, a...))
	if !strings.HasSuffix(template, "\n") {
		buildedString.WriteString("\n")
	}
	_, err := l.writeTo.Write([]byte(buildedString.String()))
	if err != nil {
		fmt.Printf("error writing to log: %s\n", err)
	}
}

func (l Logger) Debug(template string, a ...any) {
	if l.level == Debug {
		l.msg(Debug, template, a...)
	}
}

func (l Logger) Info(template string, a ...any) {
	if l.level == Info || l.level == Debug {
		l.msg(Info, template, a...)
	}
}

func (l Logger) Warning(template string, a ...any) {
	if l.level == Warning || l.level == Info || l.level == Debug {
		l.msg(Warning, template, a...)
	}
}

func (l Logger) Error(template string, a ...any) {
	if l.level == Error || l.level == Warning || l.level == Info || l.level == Debug {
		l.msg(Error, template, a...)
	}
}

func (l Logger) Log(template string, a ...any) {
	l.msg(l.level, template, a...)
}
