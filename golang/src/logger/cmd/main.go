package main

import (
	"encoding/json"
	"fmt"
	"logger"
)

var (
	// default json logger
	dl = func(log logger.Logger) logger.Logger {
		return logger.LoggerFunc(func(level int, format string, a ...interface{}) {
			bytes, _ := json.Marshal(map[string]interface{}{
				"message": fmt.Sprintf(format, a...),
				"level":   logger.LvlMap[level],
			})
			log.Log(level, string(bytes), a...)
		})
	}
)

func main() {
	var l logger.Logger

	fmt.Printf("with prefix:\n")
	l = logger.PrefixLevelLogger("prefix", logger.INFO, logger.Std)
	l.Log(logger.DEBUG, "debug log level")
	l.Log(logger.INFO, "info log level")
	l.Log(logger.WARN, "warn log level")
	l.Log(logger.ERROR, "error log level")

	fmt.Printf("\nwithout prefix:\n")
	l = logger.PrefixLevelLogger("", logger.DEBUG, logger.Std)
	l.Log(logger.DEBUG, "debug log level")
	l.Log(logger.INFO, "info log level")
	l.Log(logger.WARN, "warn log level")
	l.Log(logger.ERROR, "error log level")

	fmt.Printf("\nsystemd:\n")
	l = logger.SystemdLevelLogger(logger.LvlMapToSyslog, logger.INFO, logger.StdF)
	l.Log(logger.DEBUG, "debug log level")
	l.Log(logger.INFO, "info log level")
	l.Log(logger.WARN, "warn log level")
	l.Log(logger.ERROR, "error log level")

	fmt.Printf("\njson systemd:\n")

	l = logger.ComposerLogger(
		func(_ logger.Logger) logger.Logger { return logger.BaseLogger(logger.WARN, logger.StdF) },
		func(log logger.Logger) logger.Logger {
			return logger.SystemdLogger(logger.LvlMapToSyslog, log)
		},
		func(log logger.Logger) logger.Logger {
			return logger.JSONLogger(map[string]logger.Logger{"": dl(log)})
		},
	)
	l.Log(logger.DEBUG, "debug log level")
	l.Log(logger.INFO, "info log level")
	l.Log(logger.WARN, "warn log level")
	l.Log(logger.ERROR, "error log level")
}
