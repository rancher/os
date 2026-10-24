package logger

import (
	"fmt"
	"os"
	"strconv"

	"github.com/burmilla/os/pkg/libcompose/logger"
)

// ColorLoggerFactory implements logger.Factory interface using ColorLogger.
type ColorLoggerFactory struct {
	maxLength int
	tty       bool
}

// ColorLogger implements logger.Logger interface with color support.
type ColorLogger struct {
	name        string
	colorPrefix string
	factory     *ColorLoggerFactory
}

// NewColorLoggerFactory creates a new ColorLoggerFactory.
func NewColorLoggerFactory() *ColorLoggerFactory {
	return &ColorLoggerFactory{
		tty: isTerminal(os.Stdout),
	}
}

// Create implements logger.Factory.Create.
func (c *ColorLoggerFactory) Create(name string) logger.Logger {
	if c.maxLength < len(name) {
		c.maxLength = len(name)
	}

	return &ColorLogger{
		name:        name,
		factory:     c,
		colorPrefix: <-colorPrefix,
	}
}

// Out implements logger.Logger.Out.
func (c *ColorLogger) Out(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	logFmt, name := c.getLogFmt()
	message := fmt.Sprintf(logFmt, name, string(bytes))
	fmt.Print(message)
}

// Err implements logger.Logger.Err.
func (c *ColorLogger) Err(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	logFmt, name := c.getLogFmt()
	message := fmt.Sprintf(logFmt, name, string(bytes))
	fmt.Fprint(os.Stderr, message)
}

func (c *ColorLogger) getLogFmt() (string, string) {
	pad := c.factory.maxLength

	logFmt := "%s | %s"
	if c.factory.tty {
		logFmt = c.colorPrefix + " %s"
	}

	name := fmt.Sprintf("%-"+strconv.Itoa(pad)+"s", c.name)

	return logFmt, name
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
