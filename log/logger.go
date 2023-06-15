package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// SetLogLvl sets the log level of the logger
func SetLogLvl(l *logrus.Logger) {
	logLevel := os.Getenv("LOG_LEVEL")

	switch logLevel {
	case "DEBUG":
		l.SetLevel(logrus.DebugLevel)
	case "WARN":
		l.SetLevel(logrus.WarnLevel)
	case "INFO":
		l.SetLevel(logrus.InfoLevel)
	case "ERROR":
		l.SetLevel(logrus.ErrorLevel)
	case "TRACE":
		l.SetLevel(logrus.TraceLevel)
	case "FATAL":
		l.SetLevel(logrus.FatalLevel)
	default:
		l.SetLevel(logrus.DebugLevel)
	}
}

// SetLogType sets the log type of the logger
func SetLogType(l *logrus.Logger) {
	logType := os.Getenv("LOG_TYPE")

	switch strings.ToLower(logType) {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: true,
		})
	default:
		l.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}
