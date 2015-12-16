package log

import (
	"io"

	"fmt"

	"sync"

	"os"

	"github.com/Sirupsen/logrus"
)

type LogrusMachineLogger struct {
	history      []string
	historyLock  sync.Locker
	stdErrlogger *logrus.Logger
	stdOutlogger *logrus.Logger
}

// NewLogrusMachineLogger creates the MachineLogger implementation used by the docker-machine
func NewLogrusMachineLogger() MachineLogger {
	return &LogrusMachineLogger{
		history:      []string{},
		historyLock:  &sync.Mutex{},
		stdErrlogger: newLogger(os.Stderr),
		stdOutlogger: newLogger(os.Stdout),
	}
}

func newLogger(out io.Writer) *logrus.Logger {
	logger := logrus.New()
	logger.Level = logrus.InfoLevel
	logger.Out = out
	logger.Formatter = new(MachineFormatter)
	return logger
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func (ml *LogrusMachineLogger) RedirectStdOutToStdErr() {
	ml.stdOutlogger.Level = logrus.ErrorLevel
}

func (ml *LogrusMachineLogger) SetDebug(debug bool) {
	if debug {
		ml.stdErrlogger.Level = logrus.DebugLevel
	} else {
		ml.stdErrlogger.Level = logrus.InfoLevel
	}
}

func (ml *LogrusMachineLogger) SetErrWriter(out io.Writer) {
	ml.stdErrlogger.Out = out
}

func (ml *LogrusMachineLogger) SetOutWriter(out io.Writer) {
	ml.stdOutlogger.Out = out
}

func (ml *LogrusMachineLogger) Logger() (*logrus.Logger, *logrus.Logger) {
	return ml.stdErrlogger, ml.stdOutlogger
}

func (ml *LogrusMachineLogger) Debug(args ...interface{}) {
	ml.record(args...)
	ml.stdErrlogger.Debug(args...)
}

func (ml *LogrusMachineLogger) Debugf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.stdErrlogger.Debugf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Error(args ...interface{}) {
	ml.record(args...)
	ml.stdErrlogger.Error(args...)
}

func (ml *LogrusMachineLogger) Errorf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.stdErrlogger.Errorf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Info(args ...interface{}) {
	ml.record(args...)
	ml.stdOutlogger.Info(args...)
}

func (ml *LogrusMachineLogger) Infof(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.stdOutlogger.Infof(fmtString, args...)
}

func (ml *LogrusMachineLogger) Fatal(args ...interface{}) {
	ml.record(args...)
	ml.stdErrlogger.Fatal(args...)
}

func (ml *LogrusMachineLogger) Fatalf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.stdErrlogger.Fatalf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Warn(args ...interface{}) {
	ml.record(args...)
	ml.stdErrlogger.Warn(args...)
}

func (ml *LogrusMachineLogger) Warnf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.stdErrlogger.Warnf(fmtString, args...)
}

func (ml *LogrusMachineLogger) History() []string {
	return ml.history
}

func (ml *LogrusMachineLogger) record(args ...interface{}) {
	ml.historyLock.Lock()
	defer ml.historyLock.Unlock()
	ml.history = append(ml.history, fmt.Sprint(args...))
}

func (ml *LogrusMachineLogger) recordf(fmtString string, args ...interface{}) {
	ml.historyLock.Lock()
	defer ml.historyLock.Unlock()
	ml.history = append(ml.history, fmt.Sprintf(fmtString, args...))
}
