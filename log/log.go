package log

import (
	"io"
	"os"

	"github.com/Sirupsen/logrus"
)

// Default to using the logrus standard logger until log.InitLogger(logLevel) is called
var appLog = logrus.StandardLogger()
var userHook *ShowuserlogHook

type Fields logrus.Fields
type Level logrus.Level
type Logger logrus.Logger

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

func SetOutput(out io.Writer) {
	appLog.Out = out
}
func SetLevel(level Level) {
	if userHook != nil {
		userHook.Level = logrus.Level(level)
	} else {
		appLog.Level = logrus.Level(level)
		logrus.SetLevel(logrus.Level(level))
	}
}

func GetLevel() Level {
	if userHook != nil {
		return Level(userHook.Level)
	}
	return Level(appLog.Level)
}

func Debugf(format string, args ...interface{}) {
	appLog.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	appLog.Infof(format, args...)
}
func Printf(format string, args ...interface{}) {
	appLog.Printf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	appLog.Warnf(format, args...)
}
func Warningf(format string, args ...interface{}) {
	appLog.Warningf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	appLog.Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	appLog.Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	appLog.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	appLog.Debug(args...)
}
func Info(args ...interface{}) {
	appLog.Info(args...)
}
func Print(args ...interface{}) {
	appLog.Print(args...)
}
func Warn(args ...interface{}) {
	appLog.Warn(args...)
}
func Warning(args ...interface{}) {
	appLog.Warning(args...)
}
func Error(args ...interface{}) {
	appLog.Error(args...)
}
func Fatal(args ...interface{}) {
	appLog.Fatal(args...)
}
func Panic(args ...interface{}) {
	appLog.Panic(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return appLog.WithField(key, value)
}
func WithFields(fields Fields) *logrus.Entry {
	return appLog.WithFields(logrus.Fields(fields))
}

func InitLogger() {
	if userHook != nil {
		return // we've already initialised it
	}
	thisLog := logrus.New()

	// Filter what the user sees (info level, unless they set --debug)
	stdLogger := logrus.StandardLogger()
	showuserHook, err := NewShowuserlogHook(stdLogger.Level)
	if err != nil {
		logrus.Errorf("hook failure %s", err)
		return
	}

	filename := "/dev/kmsg"
	f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
	if err != nil {
		logrus.Debugf("error opening %s: %s", filename, err)
	} else {
		// We're all set up, now we can make it global
		appLog = thisLog
		userHook = showuserHook

		thisLog.Hooks.Add(showuserHook)
		logrus.StandardLogger().Hooks.Add(showuserHook)

		thisLog.Out = f
		logrus.SetOutput(f)
		thisLog.Level = logrus.DebugLevel
	}

	pwd, err := os.Getwd()
	if err != nil {
		thisLog.Error(err)
	}

	thisLog.Debugf("START: %v in %s", os.Args, pwd)
}
