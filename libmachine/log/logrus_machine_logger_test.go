package log

import (
	"testing"

	"bufio"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLevelIsInfo(t *testing.T) {
	errLogger, outLogger := NewLogrusMachineLogger().(*LogrusMachineLogger).Logger()
	assert.Equal(t, errLogger.Level, logrus.InfoLevel)
	assert.Equal(t, outLogger.Level, logrus.InfoLevel)
}

func TestSetDebugToTrue(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.SetDebug(true)
	errLogger, outLogger := testLogger.Logger()
	assert.Equal(t, errLogger.Level, logrus.DebugLevel)
	assert.Equal(t, outLogger.Level, logrus.InfoLevel)
}

func TestSetDebugToFalse(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.SetDebug(true)
	testLogger.SetDebug(false)
	errLogger, outLogger := testLogger.Logger()
	assert.Equal(t, errLogger.Level, logrus.InfoLevel)
	assert.Equal(t, outLogger.Level, logrus.InfoLevel)
}

func TestSetSilenceOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.RedirectStdOutToStdErr()
	errLogger, outLogger := testLogger.Logger()
	assert.Equal(t, errLogger.Level, logrus.InfoLevel)
	assert.Equal(t, outLogger.Level, logrus.ErrorLevel)
}

func TestDebugOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()
	testLogger.SetDebug(true)

	result := captureStdErr(testLogger, func() { testLogger.Debug("debug") })

	assert.Equal(t, result, "debug")
}

func TestInfoOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureStdOut(testLogger, func() { testLogger.Info("info") })

	assert.Equal(t, result, "info")
}

func TestWarnOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureStdErr(testLogger, func() { testLogger.Warn("warn") })

	assert.Equal(t, result, "warn")
}

func TestErrorOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureStdErr(testLogger, func() { testLogger.Error("error") })

	assert.Equal(t, result, "error")
}

func TestEntriesAreCollected(t *testing.T) {
	testLogger := NewLogrusMachineLogger()
	testLogger.RedirectStdOutToStdErr()
	testLogger.Debug("debug")
	testLogger.Info("info")
	testLogger.Error("error")
	assert.Equal(t, 3, len(testLogger.History()))
	assert.Equal(t, "debug", testLogger.History()[0])
	assert.Equal(t, "info", testLogger.History()[1])
	assert.Equal(t, "error", testLogger.History()[2])
}

func captureStdErr(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetErrWriter(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}

func captureStdOut(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetOutWriter(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}
