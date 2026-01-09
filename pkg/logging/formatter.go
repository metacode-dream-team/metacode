package logging

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LogField is a type of log field.
type LogField string

const (
	LogTime = LogField("time")
	Level   = LogField("level")
	Msg     = LogField("msg")
	Caller  = LogField("caller")
)

var (
	defaultDelimiter  = " || "
	defaultLogFields  = []LogField{LogTime, Level, Msg}
	defaultTimeFormat = time.RFC3339
)

// Formatter implements logrus.Formatter
type Formatter struct {
	Delimiter  string
	LogFields  []LogField
	TimeFormat string
}

// Format formats the log entry
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.setDefaultValues()

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	for i, field := range f.LogFields {
		if i > 0 {
			b.WriteString(f.Delimiter)
		}
		switch field {
		case LogTime:
			b.WriteString(entry.Time.Format(f.TimeFormat))
		case Level:
			levelStr := strings.ToUpper(entry.Level.String())
			color := getLevelColor(entry.Level)
			b.WriteString(fmt.Sprintf("%s%s%s", color, levelStr, Reset))
		case Msg:
			b.WriteString(entry.Message)
			if entry.Data != nil {
				for k, v := range entry.Data {
					fmt.Fprintf(b, "%s%s=%v", f.Delimiter, k, v)
				}
			}
		case Caller:
			if entry.HasCaller() {
				fmt.Fprintf(b, "%s:%d", entry.Caller.File, entry.Caller.Line)
			}
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// Assign colors based on log level
func getLevelColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return Cyan
	case logrus.InfoLevel:
		return Green
	case logrus.WarnLevel:
		return Yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return Red
	default:
		return White
	}
}

func (f *Formatter) setDefaultValues() {
	if f.Delimiter == "" {
		f.Delimiter = defaultDelimiter
	}
	if f.LogFields == nil || len(f.LogFields) == 0 {
		f.LogFields = defaultLogFields
	}
	if f.TimeFormat == "" {
		f.TimeFormat = defaultTimeFormat
	}
}
