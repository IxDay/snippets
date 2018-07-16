package logger

import (
	"io"
	"log"
	"log/syslog"
	"os"
	"strconv"
)

const (
	DEBUG int = iota + 1
	INFO
	WARN
	ERROR
	OFF
)

var (
	LvlMap = map[int]string{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		OFF:   "",
	}
	LvlMapToSyslog = PriorityMap{
		DEBUG: syslog.LOG_DEBUG,
		INFO:  syslog.LOG_INFO,
		WARN:  syslog.LOG_WARNING,
		ERROR: syslog.LOG_ERR,
	}
	Std         = log.New(os.Stderr, "", log.LstdFlags)
	StdF        = log.New(os.Stderr, "", 0)
	LogPatchers = []io.Writer{}
)

type (
	PriorityMap map[int]syslog.Priority
	WriterFunc  func(p []byte) (n int, err error)
	LoggerFunc  func(int, string, ...interface{})
	Logger      interface {
		Log(level int, format string, a ...interface{})
	}
)

func init() {
	// hot patch log
	log.SetFlags(0)
	log.SetOutput(WriterFunc(func(p []byte) (int, error) {
		return io.MultiWriter(LogPatchers...).Write(p)
	}))
}

func LoggerWriter(mark int, logger Logger) io.Writer {
	return WriterFunc(func(p []byte) (int, error) {
		logger.Log(mark, "%s", p)
		return len(p), nil
	})
}

func StdLogger(mark int, logger Logger) *log.Logger {
	return log.New(LoggerWriter(mark, logger), "", 0)
}

func (wf WriterFunc) Write(p []byte) (n int, err error) { return wf(p) }
func (lf LoggerFunc) Log(level int, format string, a ...interface{}) {
	lf(level, format, a...)
}

func NoopLogger() Logger {
	return LoggerFunc(func(_ int, _ string, _ ...interface{}) {})
}

func BaseLogger(mark int, logger *log.Logger) Logger {
	return LoggerFunc(func(level int, format string, a ...interface{}) {
		if level >= mark {
			logger.Printf(format, a...)
		}
	})
}

func LevelLogger(logger Logger) Logger {
	return LoggerFunc(func(level int, format string, a ...interface{}) {
		logger.Log(level, "["+LvlMap[level]+"] "+format, a...)
	})
}

func PrefixLogger(prefix string, logger Logger) Logger {
	if prefix != "" {
		prefix = prefix + ": "
	}
	return LoggerFunc(func(level int, format string, a ...interface{}) {
		logger.Log(level, prefix+format, a...)
	})
}

func SystemdLogger(lvlMap PriorityMap, logger Logger) Logger {
	return LoggerFunc(func(level int, format string, a ...interface{}) {
		logger.Log(level, "<"+strconv.Itoa(int(lvlMap[level]))+">"+format, a...)
	})
}

func JSONLogger(loggers map[string]Logger) Logger {
	return LoggerFunc(func(level int, format string, a ...interface{}) {
		if logger, ok := loggers[format]; ok {
			logger.Log(level, format, a...)
		} else if logger, ok := loggers[""]; ok {
			// set a default
			logger.Log(level, format, a...)
		} else {
			// do something if nothing, maybe fallback on something or panic or whatever
		}
	})
}

func ComposerLogger(loggers ...func(Logger) Logger) Logger {
	base := NoopLogger()
	for _, logger := range loggers {
		base = logger(base)
	}
	return base
}

func PrefixLevelLogger(prefix string, mark int, logger *log.Logger) Logger {
	return ComposerLogger(
		func(_ Logger) Logger { return BaseLogger(mark, logger) },
		func(logger Logger) Logger { return LevelLogger(logger) },
		func(logger Logger) Logger { return PrefixLogger(prefix, logger) },
	)
}

func SystemdLevelLogger(lvlMap PriorityMap, mark int, logger *log.Logger) Logger {
	return ComposerLogger(
		func(_ Logger) Logger { return BaseLogger(mark, logger) },
		func(logger Logger) Logger { return SystemdLogger(lvlMap, logger) },
	)
}
