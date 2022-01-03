package logFormatter

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

const defaultTimestampFormat = "2006-01-02T15:04:05.000"

// FieldinFormatter formats logs into text the way fieldin wants to
type FieldinFormatter struct {

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed.
	// The format to use is the same than for time.Format or time.Parse from the standard
	// library.
	// The standard Library already provides a set of predefined format.
	TimestampFormat string
}

// Format renders a single log entry
func (f *FieldinFormatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(log.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	prefixFieldClashes(data, entry.HasCaller())
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	var funcVal, fileVal string

	fixedKeys := make([]string, 0, 4+len(data))
	if !f.DisableTimestamp {
		fixedKeys = append(fixedKeys, log.FieldKeyTime)
	}
	fixedKeys = append(fixedKeys, log.FieldKeyLevel)
	if entry.Message != "" {
		fixedKeys = append(fixedKeys, log.FieldKeyMsg)
	}
	if err, ok := data[log.FieldKeyLogrusError]; ok && err != "" {
		fixedKeys = append(fixedKeys, log.FieldKeyLogrusError)
	}

	if entry.HasCaller() {
		funcVal = entry.Caller.Function
		fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)

		if funcVal != "" {
			fixedKeys = append(fixedKeys, log.FieldKeyFunc)
		}
		if fileVal != "" {
			fixedKeys = append(fixedKeys, log.FieldKeyFile)
		}
	}

	fixedKeys = append(fixedKeys, keys...)

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}

	f.printLogLine(b, entry, keys, data, timestampFormat)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *FieldinFormatter) printLogLine(b *bytes.Buffer, entry *log.Entry, keys []string, data log.Fields, timestampFormat string) {

	levelText := strings.ToUpper(entry.Level.String())

	// Remove a single newline if it already exists in the message to keep
	// the behavior of logrus text_formatter the same as the stdlib log package
	entry.Message = strings.TrimSuffix(entry.Message, "\n")

	caller := ""
	if entry.HasCaller() {
		funcVal := fmt.Sprintf("%s()", entry.Caller.Function)
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)

		if fileVal == "" {
			caller = funcVal
		} else if funcVal == "" {
			caller = fileVal
		} else {
			caller = fileVal + " " + funcVal
		}
	}

	fmt.Fprintf(b, "%s %s", entry.Time.Format(timestampFormat), levelText)
	if entry.HasCaller() && caller != "" {
		fmt.Fprintf(b, " %s", caller)
	}
	addSemiColon := false
	for _, k := range keys {
		addSemiColon = true
		v := data[k]
		fmt.Fprintf(b, " %s=", k)
		f.appendValue(b, v)
	}
	if addSemiColon {
		fmt.Fprint(b, " ;")
	}
	if entry.Message != "" {
		fmt.Fprintf(b, " %s", entry.Message)
	}
}

func (f *FieldinFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *FieldinFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	b.WriteString(fmt.Sprintf("%q", stringVal))
}

// This is to not silently overwrite `time`, `msg`, `func` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//  logrus.WithField("level", 1).Info("hello")
//
// Would just silently drop the user provided level. Instead with this code
// it'll logged as:
//
//  {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
//
func prefixFieldClashes(data log.Fields, reportCaller bool) {
	timeKey := log.FieldKeyTime
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}

	msgKey := log.FieldKeyMsg
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}

	levelKey := log.FieldKeyLevel
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}

	logrusErrKey := log.FieldKeyLogrusError
	if l, ok := data[logrusErrKey]; ok {
		data["fields."+logrusErrKey] = l
		delete(data, logrusErrKey)
	}

	// If reportCaller is not set, 'func' will not conflict.
	if reportCaller {
		funcKey := log.FieldKeyFunc
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := log.FieldKeyFile
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
}
