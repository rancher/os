package log

import (
	"io"
	"os"
	"path/filepath"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

var userHook *ShowuserlogHook
var defaultLogLevel logrus.Level
var debugThisLogger = false

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
	logrus.SetOutput(out)
}
func SetDefaultLevel(level Level) {
	defaultLogLevel = logrus.Level(level)
}
func SetLevel(level Level) {
	if userHook != nil {
		userHook.Level = logrus.Level(level)
	} else {
		logrus.SetLevel(logrus.Level(level))
	}
}

func GetLevel() Level {
	if userHook != nil {
		return Level(userHook.Level)
	}
	return Level(logrus.GetLevel())
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}
func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}
func Warningf(format string, args ...interface{}) {
	logrus.Warningf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}
func Info(args ...interface{}) {
	logrus.Info(args...)
}
func Print(args ...interface{}) {
	logrus.Print(args...)
}
func Warn(args ...interface{}) {
	logrus.Warn(args...)
}
func Warning(args ...interface{}) {
	logrus.Warning(args...)
}
func Error(args ...interface{}) {
	logrus.Error(args...)
}
func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}
func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}
func WithFields(fields Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(fields))
}

// InitLogger sets up Logging to log to /dev/kmsg and to Syslog
func InitLogger() {
	if logTheseApps() {
		innerInit(false)
		FsReady()

		pwd, err := os.Getwd()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debugf("START: %v in %s", os.Args, pwd)
	}
}

func logTheseApps() bool {
	// TODO: mmm, not very functional.
	if filepath.Base(os.Args[0]) == "ros" ||
		filepath.Base(os.Args[0]) == "host_ros" ||
		filepath.Base(os.Args[0]) == "system-docker" {
		return false
	}
	return true
}

// InitDeferedLogger stores the log messages until FsReady() is called
// TODO: actually store them :)
// TODO: need to work out how to pass entries from a binary run before we switchfs back to init and have it store and write it later
func InitDeferedLogger() {
	if logTheseApps() {
		innerInit(true)
		logrus.SetOutput(ioutil.Discard)

		pwd, err := os.Getwd()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debugf("START: %v in %s", os.Args, pwd)
	}
}

func innerInit(deferedHook bool) {
	if userHook != nil {
		return // we've already initialised it
	}

	// All logs go through the Hooks, and they choose what to do with them.
	logrus.StandardLogger().Level = logrus.DebugLevel

	if logTheseApps() {
		AddUserHook(deferedHook)
	}
}

func FsReady() {
	filename := "/var/log/boot/" + filepath.Base(os.Args[0]) + ".log"
	if err := os.MkdirAll(filepath.Dir(filename), os.ModeDir|0755); debugThisLogger && err != nil {
		logrus.Errorf("FsReady mkdir(%s): %s", filename, err)
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		if debugThisLogger {
			logrus.Errorf("FsReady opening %s: %s", filename, err)
		}
	} else {
		if debugThisLogger {
			logrus.Infof("Setting log output for %s to: %s", os.Args[0], filename)
		}
		logrus.SetOutput(f)
	}
}

// AddUserHook is used to filter what log messages are written to the screen
func AddUserHook(deferedHook bool) error {
	if userHook != nil {
		return nil
	}

	printLogLevel := logrus.InfoLevel

	uh, err := NewShowuserlogHook(printLogLevel, filepath.Base(os.Args[0]))
	if err != nil {
		logrus.Errorf("error creating userHook(%s): %s", os.Args[0], err)
		return err
	}
	userHook = uh
	logrus.StandardLogger().Hooks.Add(uh)

	if debugThisLogger {
		if deferedHook {
			logrus.Debugf("------------info Starting defered User Hook (%s)", os.Args[0])
		} else {
			logrus.Debugf("------------info Starting User Hook (%s)", os.Args[0])
		}
	}

	return nil
}
