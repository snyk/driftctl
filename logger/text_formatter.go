package logger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/sirupsen/logrus"
)

var baseTimestamp time.Time

func init() {
	baseTimestamp = time.Now()
}

// TextFormatter formats logs into text
type TextFormatter struct {
	// The max length of the level text, generated dynamically on init if == 0
	levelTextMaxLength int
}

func NewTextFormatter(levelTextMaxLength int) *TextFormatter {
	if levelTextMaxLength <= 0 {
		for _, level := range logrus.AllLevels {
			levelLen := len(level.String())
			if levelLen > levelTextMaxLength {
				levelTextMaxLength = levelLen
			}
		}
	}

	return &TextFormatter{levelTextMaxLength: levelTextMaxLength}
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if err := f.writeLevel(entry, b); err != nil {
		return nil, err
	}

	if err := f.writeElapsedTime(entry, b); err != nil {
		return nil, err
	}

	if err := f.writeMessage(entry, b); err != nil {
		return nil, err
	}

	if err := f.writeContext(entry, b); err != nil {
		return nil, err
	}

	if err := f.writeCaller(entry, b); err != nil {
		return nil, err
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) writeCaller(entry *logrus.Entry, b *bytes.Buffer) error {
	if entry.HasCaller() {
		caller := ""

		funcVal := fmt.Sprintf("%s()", entry.Caller.Function)
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)

		if fileVal == "" {
			caller = caller + funcVal
		} else if funcVal == "" {
			caller = fileVal
		} else {
			caller = fileVal + " " + funcVal
		}

		if _, err := fmt.Fprintf(b, " (%s)", caller); err != nil {
			return err
		}
	}
	return nil
}

func (f *TextFormatter) writeContext(entry *logrus.Entry, b *bytes.Buffer) error {
	keys := make([]string, 0)
	for key := range entry.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if _, err := fmt.Fprintf(b, " %s=%s", color.CyanString("%s", key), entry.Data[key]); err != nil {
			return err
		}
	}
	return nil
}

func (f *TextFormatter) writeMessage(entry *logrus.Entry, b *bytes.Buffer) error {
	if _, err := color.New(color.FgHiWhite).Fprintf(b, " %s", entry.Message); err != nil {
		return err
	}
	return nil
}

func (f *TextFormatter) writeElapsedTime(entry *logrus.Entry, b *bytes.Buffer) error {
	if _, err := fmt.Fprintf(b, "[%04d]", int(entry.Time.Sub(baseTimestamp)/time.Second)); err != nil {
		return err
	}
	return nil
}

func (f *TextFormatter) writeLevel(entry *logrus.Entry, b *bytes.Buffer) error {
	levelText := strings.ToUpper(entry.Level.String())

	var levelColor *color.Color
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = color.New(color.Bold, color.FgHiWhite)
	case logrus.WarnLevel:
		levelColor = color.New(color.Bold, color.FgYellow)
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = color.New(color.Bold, color.FgRed)
	default:
		levelColor = color.New(color.Bold, color.FgBlue)
	}

	if len(levelText) > f.levelTextMaxLength {
		levelText = levelText[0:f.levelTextMaxLength] // TRUNCATE if needed
	}
	// and then pad to f.levelTextMaxLength
	if _, err := levelColor.Fprintf(b, "%*v", -f.levelTextMaxLength, levelText); err != nil {
		return err
	}
	return nil
}
