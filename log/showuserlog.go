package log

import (
	"fmt"
	"log/syslog"
	"os"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
)

// ShowuserlogHook stores all levels of logrus entries in memory until its told the RancherOS logging system is ready
// then it replays them to be logged
type ShowuserlogHook struct {
	Level         logrus.Level
	syslogHook    *logrus_syslog.SyslogHook
	storedEntries []*logrus.Entry
}

// NewShowuserlogHook creates a new hook for use
func NewShowuserlogHook(l logrus.Level) (*ShowuserlogHook, error) {
	return &ShowuserlogHook{l, nil, []*logrus.Entry{}}, nil
}

// Fire is called by logrus when the Hook is active
func (hook *ShowuserlogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	if entry.Level <= hook.Level {
		fmt.Printf("SVEN %s", line)
	}

	if hook.syslogHook == nil {
		hook.storedEntries = append(hook.storedEntries, entry)
	} else {
		err := hook.syslogHook.Fire(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOGERR: Unable to write %s, %v", line, err)
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

// Set up Syslog Hook, and replay any stored entries.
func (hook *ShowuserlogHook) LogSystemReady() error {
	if hook.syslogHook == nil {
		h, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
		if err != nil {
			logrus.Debugf("error creating SyslogHook: %s", err)
			return err
		}
		hook.syslogHook = h

		for _, entry := range hook.storedEntries {
			hook.syslogHook.Fire(entry)
		}
	}

	return nil
}
