package log

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// This implementation for log is ***typically*** used for surfacing events that don't have an exit flow.
// (meaning for things like stream and state interfaces who's flow ends when the invocation is done)
// In the cases of errors, logging using this package is only acceptable when it doesn't make sense to
// return the error from the function.

// TL;DR: use this for invocation errors, DON'T use this for semantic/context errors.

type LogLevel string

func (level LogLevel) Index() int {
	for i, val := range LoggerLevels {
		if val == level {
			return i
		}
	}
	return -1
}

const (
	LevelDEBUG LogLevel = "DEBUG"
	LevelINFO  LogLevel = "INFO"
	LevelWARN  LogLevel = "WARN"
	LevelERROR LogLevel = "ERROR"
	LevelFATAL LogLevel = "FATAL"
)

var LoggerLevels = [...]LogLevel{LevelDEBUG, LevelINFO, LevelWARN, LevelERROR, LevelFATAL}

type Logger interface {
	Log(msg LoggerMessage)
}

var LogErrorSignal = Signal("LOG_ERROR")

type LogData map[string]interface{}

type LoggerMessage struct {
	Timestamp time.Time   `json:"timestamp"`
	LogLevel  LogLevel    `json:"level"`
	Signal    *string     `json:"signal,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

var current Logger = StandardLogger{
	out: os.Stdout,
	err: os.Stderr,
}

func SetLogger(l Logger) {
	current = l
}

var currentLevel int

func init() {
	osLevel := os.Getenv("LOG_LEVEL")
	if osLevel != "" {
		currentLevel = LogLevel(strings.ToUpper(osLevel)).Index()
		return
	}
	currentLevel = LevelINFO.Index()
}

func SetLevel(level LogLevel) {
	for i, val := range LoggerLevels {
		if val == level {
			currentLevel = i
			return
		}
	}
	currentLevel = 1
}

func Signal(s string) *string {
	return &s
}

func Output(msg LoggerMessage) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	levelIdx := -1
	for i, val := range LoggerLevels {
		if val == msg.LogLevel {
			levelIdx = i
		}
	}
	if levelIdx == -1 {
		msg.LogLevel = LevelINFO
		levelIdx = 1
	}
	if currentLevel <= levelIdx {
		current.Log(msg)
	}
}

func Print(level LogLevel, signal *string, message string) {
	Output(LoggerMessage{
		Timestamp: time.Now(),
		LogLevel:  level,
		Signal:    signal,
		Message:   message,
		Data:      nil,
	})
}
func Printf(level LogLevel, signal *string, format string, v ...any) {
	Output(LoggerMessage{
		Timestamp: time.Now(),
		LogLevel:  level,
		Signal:    signal,
		Message:   fmt.Sprintf(format, v...),
		Data:      nil,
	})
}
