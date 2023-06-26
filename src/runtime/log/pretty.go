package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type PrettyLogger struct {
	out io.Writer
	err io.Writer
}

var ttyTimestamp = color.New(color.FgCyan).SprintFunc()
var ttySignal = color.New(color.Bold, color.FgWhite).SprintFunc()

var ttyDebugTag = color.New(color.BgBlue, color.FgWhite).Sprint(" DEBUG ")
var ttyInfoTag = color.New(color.BgGreen, color.FgWhite).Sprint(" INFO ")
var ttyWarnTag = color.New(color.BgYellow, color.FgBlack).Sprint(" WARN ")
var ttyErrorTag = color.New(color.BgRed, color.FgWhite).Sprint(" ERROR ")
var ttyFatalTag = color.New(color.BgMagenta, color.FgWhite).Sprint(" FATAL ")
var ttyDefaultTagFn = color.New(color.BgWhite, color.FgBlack).SprintFunc()

func NewPrettyLogger() PrettyLogger {
	return PrettyLogger{
		out: os.Stdout,
		err: os.Stderr,
	}
}

func normalizeDigits(num int) string {
	if num < 10 {
		return fmt.Sprintf("0%v", num)
	}
	return fmt.Sprint(num)
}

func (l PrettyLogger) Log(msg LoggerMessage) {
	out := ttyTimestamp(
		fmt.Sprintf(
			"[%v:%v:%v] ",
			msg.Timestamp.Hour(),
			normalizeDigits(msg.Timestamp.Minute()),
			normalizeDigits(msg.Timestamp.Second()),
		),
	)

	switch msg.LogLevel {
	case LevelDEBUG:
		out += ttyDebugTag + " "
	case LevelINFO:
		out += ttyInfoTag + " "
	case LevelWARN:
		out += ttyWarnTag + " "
	case LevelERROR:
		out += ttyErrorTag + " "
	case LevelFATAL:
		out += ttyFatalTag + " "
	default:
		out += ttyDefaultTagFn(fmt.Sprintf(" %s ", msg.LogLevel)) + " "
	}

	if msg.Signal != nil {
		out += fmt.Sprintf("%s: ", ttySignal(*msg.Signal))
	}
	out += msg.Message

	if msg.Data != nil {
		bytes, err := json.Marshal(msg.Data)
		if err != nil {
			Output(LoggerMessage{
				LogLevel: LevelERROR,
				Signal:   LogErrorSignal,
				Message:  "failed to marshal data",
				Data:     fmt.Sprintf("%v", msg),
			})
			return
		}
		out += fmt.Sprintf("\n-- %s", string(bytes))
	}

	bytes := append([]byte(out), byte(10))
	if msg.LogLevel.Index() >= LevelERROR.Index() {
		l.err.Write(bytes)
		return
	}
	l.out.Write(bytes)
}
