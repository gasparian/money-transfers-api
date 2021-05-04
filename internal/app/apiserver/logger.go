package apiserver

import (
	"log"
	"os"
)

const (
	errorLevel = iota
	warnLevel
	infoLevel
	debugLevel
)

var (
	levels = map[string]int{
		"error":   0,
		"warning": 1,
		"info":    2,
		"debug":   3,
	}
)

type logger struct {
	level int
	warn  *log.Logger
	info  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

// NewLogger holds stdlib loggers for every needed logging level
func NewLogger() *logger {
	return &logger{
		level: levels["info"],
		err:   log.New(os.Stderr, "[ Error ] ", log.LstdFlags|log.Lshortfile),
		warn:  log.New(os.Stdout, "[ Warn ] ", log.LstdFlags|log.Lshortfile),
		info:  log.New(os.Stdout, "[ Info ] ", log.LstdFlags|log.Lshortfile),
		debug: log.New(os.Stderr, "[ Debug ] ", log.LstdFlags|log.Lshortfile),
	}
}

// SetLevel sets needed logging level
func (l *logger) SetLevel(level string) {
	l.level = levels[level]
}

// Error writes out message based on current logging level
func (l *logger) Error(msg string) {
	if l.level >= errorLevel {
		l.err.Println(msg)
	}
}

// Warn ...
func (l *logger) Warn(msg string) {
	if l.level >= warnLevel {
		l.warn.Println(msg)
	}
}

// Info ...
func (l *logger) Info(msg string) {
	if l.level >= infoLevel {
		l.info.Println(msg)
	}
}

// Debug ...
func (l *logger) Debug(msg string) {
	if l.level >= debugLevel {
		l.debug.Println(msg)
	}
}
