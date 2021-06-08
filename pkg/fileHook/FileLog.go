// +build windows

package filehook

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

// FileHook to send logs via windows log.
type FileHook struct {
	path string
	file *os.File
}

// NewHook creates and returns a new FileHook wrapped around anything that implements the debug.Log interface
func NewHook(path string) *FileHook {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	hook := FileHook{path: path, file: file}
	return &hook
}
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	_, err := hook.file.WriteString(fmt.Sprintf("[%s] %v: %s\r\n", entry.Level.String(), entry.Time.String(), entry.Message))
	return err
}

func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
