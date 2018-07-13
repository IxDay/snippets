package logger

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func helperNewLogger(t *testing.T) (*log.Logger, *bytes.Buffer) {
	t.Helper()
	b := bytes.Buffer{}
	return log.New(&b, "", 0), &b
}

func TestPrefixLevelLogger(t *testing.T) {
	logger, buffer := helperNewLogger(t)
	pll := PrefixLevelLogger("foo", WARN, logger)

	cases := []struct {
		inLvl       int
		inStr, want string
	}{
		{WARN, "bar", "[WARN] foo: bar\n"},
		{ERROR, "bar", "[ERROR] foo: bar\n"},
		{INFO, "bar", ""},
		{DEBUG, "bar", ""},
	}
	for _, c := range cases {
		pll.Log(c.inLvl, "%s", c.inStr)
		if got := buffer.String(); got != c.want {
			t.Errorf("Log(%s, %q, %q) == %q, want %q",
				LvlMap[c.inLvl], "%s", c.inStr, got, c.want)
		}
		buffer.Reset()
	}
}

func TestSystemdLogger(t *testing.T) {
	logger, buffer := helperNewLogger(t)
	sl := SystemdLevelLogger(LvlMapToSyslog, INFO, logger)

	cases := []struct {
		inLvl       int
		inStr, want string
	}{
		{DEBUG, "foo", ""},
		{INFO, "foo", "<6>foo\n"},
		{WARN, "bar", "<4>bar\n"},
		{ERROR, "bar", "<3>bar\n"},
	}
	for _, c := range cases {
		sl.Log(c.inLvl, "%s", c.inStr)
		if got := buffer.String(); got != c.want {
			t.Errorf("SystemdLogger.Log(%s, %q, %q) == %q, want %q",
				LvlMap[c.inLvl], "%s", c.inStr, got, c.want)
		}
		buffer.Reset()
	}
}

func TestJSONLogger(t *testing.T) {
	logger, buffer := helperNewLogger(t)
	dl := LoggerFunc(func(level int, format string, _ ...interface{}) {
		format = fmt.Sprintf("{%q: %q, %q: %q}", "msg", format, "lvl", LvlMap[level])
		BaseLogger(WARN, logger).Log(level, format)
	})
	sl := JSONLogger(map[string]Logger{"bar": dl, "": dl})

	cases := []struct {
		inLvl       int
		inStr, want string
	}{
		{DEBUG, "foo", ""},
		{INFO, "bar", ""},
		{WARN, "bar", `{"msg": "bar", "lvl": "WARN"}` + "\n"},
		{ERROR, "foo", `{"msg": "foo", "lvl": "ERROR"}` + "\n"},
	}
	for _, c := range cases {
		sl.Log(c.inLvl, c.inStr)
		if got := buffer.String(); got != c.want {
			t.Errorf("SystemdLogger.Log(%s, %q, <nil>) == %s, want %s",
				LvlMap[c.inLvl], c.inStr, got, c.want)
		}
		buffer.Reset()
	}
}

func TestLogPatcher(t *testing.T) {
	logger, buffer := helperNewLogger(t)
	ll := LevelLogger(BaseLogger(ERROR, logger))
	LogPatchers = append(LogPatchers, WriterFunc(func(p []byte) (int, error) {
		ll.Log(ERROR, string(p))
		return len(p), nil
	}))

	cases := []struct{ inStr, want string }{
		{"foo", "[ERROR] foo\n"},
		{"bar", "[ERROR] bar\n"},
	}
	for _, c := range cases {
		log.Println(c.inStr)
		if got := buffer.String(); got != c.want {
			t.Errorf("log.Printf(%q) == %q, want %q", c.inStr, got, c.want)
		}
		buffer.Reset()
	}
}
