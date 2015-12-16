package log

import (
	"log"
	"os"

	"github.com/Sirupsen/logrus"
)

// FileHook handle writing to a local log file.
type FileHook struct {
	path string
}

func NewFileHook(path string) *FileHook {
	hook := &FileHook{
		path: path,
	}
	return hook
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	fd, err := os.OpenFile(hook.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("failed to open logfile:", hook.path, err)
		return err
	}
	defer fd.Close()
	msg, err := entry.String()
	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}
	fd.WriteString(msg)
	return nil
}

func (hook *FileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
