// Package types
package types

import (
	"log"
	"os"
)

type LogType struct {
	Report *log.Logger
	Error  *log.Logger
	Warn   *log.Logger
	Info   *log.Logger
	Debug  *log.Logger
	Stats  *log.Logger
}

func (l *LogType) EnableInfo() {
	l.Info.SetOutput(os.Stderr)
}

func (l *LogType) EnableDebug() {
	l.Debug.SetOutput(os.Stderr)
}

func (l *LogType) EnableStats(file *os.File) {
	l.Stats.SetOutput(file)
}
