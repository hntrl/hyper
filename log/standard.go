package log

import (
	"encoding/json"
	"fmt"
	"io"
)

type StandardLogger struct {
	out io.Writer
	err io.Writer
}

func (l StandardLogger) Log(msg LoggerMessage) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		Output(LoggerMessage{
			LogLevel: LevelERROR,
			Signal:   LogErrorSignal,
			Message:  "failed to marshal message",
			Data:     fmt.Sprintf("%v", msg),
		})
		return
	}
	bytes = append(bytes, byte(10))
	if msg.LogLevel.Index() >= LevelERROR.Index() {
		l.err.Write(bytes)
		return
	}
	l.out.Write(bytes)
}
