package log

import (
	"fmt"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

// ShowuserlogHook stores all levels of logrus entries in memory until its told the BurmillaOS logging system is ready
// then it replays them to be logged
type ShowuserlogHook struct {
	Level         logrus.Level
	syslogHook    *logrus_syslog.SyslogHook
	storedEntries []*logrus.Entry
	appName       string
}

// NewShowuserlogHook creates a new hook for use
func NewShowuserlogHook(l logrus.Level, app string) (*ShowuserlogHook, error) {
	return &ShowuserlogHook{l, nil, []*logrus.Entry{}, app}, nil
}

// Fire is called by logrus when the Hook is active
func (hook *ShowuserlogHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= hook.Level {
		//if f, err := os.OpenFile("/dev/kmsg", os.O_WRONLY, 0644); err != nil {
		//	fmt.Fprintf(f, "%s:%s: %s\n", hook.appName, entry.Level, entry.Message)
		//	f.Close()
		//} else {
		fmt.Printf("[            ] %s:%s: %s\n", hook.appName, entry.Level, entry.Message)
		//}
	}

	if hook.syslogHook == nil {
		hook.storedEntries = append(hook.storedEntries, entry)
	} else {
		err := hook.syslogHook.Fire(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOGERR: Unable to syslog.Fire %v, %v", entry, err)
			return err
		}
	}

	return nil
}

// Levels returns all log levels, so we can process them ourselves
func (hook *ShowuserlogHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

// NotUsedYetLogSystemReady Set up Syslog Hook, and replay any stored entries.
func (hook *ShowuserlogHook) NotUsedYetLogSystemReady() error {
	if hook.syslogHook == nil {
		h, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
		if err != nil {
			logrus.Debugf("error creating SyslogHook: %v", err)
			return err
		}
		hook.syslogHook = h

		for _, entry := range hook.storedEntries {
			line, _ := entry.String()
			fmt.Printf("---- CATCHUP %s\n", line)
			hook.syslogHook.Fire(entry)
		}
	}

	return nil
}
