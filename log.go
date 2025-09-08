package main

import (
	"fmt"
	"strings"
)

// Log levels
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
)

var logLevelMap = map[string]int{
	"debug": LogDebug,
	"info":  LogInfo,
	"warn":  LogWarn,
	"error": LogError,
}

// Logger provides logging functionality with configurable level
type Logger struct {
	level int
}

// NewLogger creates a new logger with the specified level
func NewLogger(levelStr string) *Logger {
	level, ok := logLevelMap[strings.ToLower(levelStr)]
	if !ok {
		level = LogInfo // Default to info if invalid
		fmt.Printf("[warn] Invalid log level '%s', defaulting to 'info'\n", levelStr)
	}
	return &Logger{level: level}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogDebug {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogInfo {
		fmt.Printf("[info] "+format+"\n", args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogWarn {
		fmt.Printf("[warn] "+format+"\n", args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogError {
		fmt.Printf("[error] "+format+"\n", args...)
	}
}
