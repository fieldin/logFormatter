package logFormatter

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestFormatting(t *testing.T) {
	tf := &FieldinFormatter{}

	testCases := []struct {
		value    string
		expected string
	}{
		{`foo`, "0001-01-01T00:00:00.000 PANIC test=\"foo\" ;\n"},
		{`foo and bar`, "0001-01-01T00:00:00.000 PANIC test=\"foo and bar\" ;\n"},
	}

	for _, tc := range testCases {
		b, _ := tf.Format(log.WithField("test", tc.value))

		if string(b) != tc.expected {
			t.Errorf("formatting expected for %q (result was %q instead of %q)", tc.value, string(b), tc.expected)
		}
	}
}

func TestQuoting(t *testing.T) {
	tf := &FieldinFormatter{}

	checkQuoting := func(q bool, value interface{}) {
		b, _ := tf.Format(log.WithField("test", value))
		idx := bytes.Index(b, ([]byte)("test="))
		cont := bytes.Contains(b[idx+5:], []byte("\""))
		if cont != q {
			if q {
				t.Errorf("quoting expected for: %#v", value)
			} else {
				t.Errorf("quoting not expected for: %#v", value)
			}
		}
	}

	checkQuoting(true, "")
	checkQuoting(true, "abcd")
	checkQuoting(true, "foo\n\rbar")
	checkQuoting(true, errors.New("invalid argument"))

}

func TestEscaping(t *testing.T) {
	tf := &FieldinFormatter{}

	testCases := []struct {
		value    string
		expected string
	}{
		{`ba"r`, `ba\"r`},
		{`ba'r`, `ba'r`},
	}

	for _, tc := range testCases {
		b, _ := tf.Format(log.WithField("test", tc.value))
		if !bytes.Contains(b, []byte(tc.expected)) {
			t.Errorf("escaping expected for %q (result was %q instead of %q)", tc.value, string(b), tc.expected)
		}
	}
}

func TestEscaping_Interface(t *testing.T) {
	tf := &FieldinFormatter{}

	ts := time.Now()

	testCases := []struct {
		value    interface{}
		expected string
	}{
		{ts, fmt.Sprintf("\"%s\"", ts.String())},
		{errors.New("error: something went wrong"), "\"error: something went wrong\""},
	}

	for _, tc := range testCases {
		b, _ := tf.Format(log.WithField("test", tc.value))
		if !bytes.Contains(b, []byte(tc.expected)) {
			t.Errorf("escaping expected for %q (result was %q instead of %q)", tc.value, string(b), tc.expected)
		}
	}
}

func TestTimestampFormat(t *testing.T) {
	checkTimeStr := func(format string) {
		customFormatter := &FieldinFormatter{TimestampFormat: format}
		customStr, _ := customFormatter.Format(log.WithField("test", "test"))
		timeStart := 0
		timeEnd := bytes.Index(customStr, ([]byte)(" PANIC"))
		timeStr := customStr[timeStart:timeEnd]
		if format == "" {
			format = defaultTimestampFormat
		}
		_, e := time.Parse(format, (string)(timeStr))
		if e != nil {
			t.Errorf("time string \"%s\" did not match provided time format \"%s\": %s", timeStr, format, e)
		}
	}

	checkTimeStr("2006-01-02T15:04:05.000000000Z07:00")
	checkTimeStr("Mon Jan _2 15:04:05 2006")
	checkTimeStr("")
}
