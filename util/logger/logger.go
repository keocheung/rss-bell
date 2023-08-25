// Package logger prints logs with different levels
package logger

import (
	"log"
	"os"
)

type level uint32

const (
	levelConfigEnvKey       = "LOG_LEVEL"
	levelDebug        level = iota
	levelInfo
	levelWarn
	levelError
)

var logLevel = levelInfo

// init sets log level from environment
func init() {
	levelConfigStr := os.Getenv(levelConfigEnvKey)
	if levelConfig, ok := levelMap[levelConfigStr]; ok {
		logLevel = levelConfig
	}
}

var levelMap = map[string]level{
	"DEBUG": levelDebug,
	"INFO":  levelInfo,
	"WARN":  levelWarn,
	"ERROR": levelError,
}

// Debugf prints logs with DEBUG level
func Debugf(format string, args ...interface{}) {
	if !shouldPrint(levelDebug) {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

// Infof prints logs with INFO level
func Infof(format string, args ...interface{}) {
	if !shouldPrint(levelInfo) {
		return
	}
	log.Printf("[INFO] "+format, args...)
}

// Warnf prints logs with WARN level
func Warnf(format string, args ...interface{}) {
	if !shouldPrint(levelWarn) {
		return
	}
	log.Printf("[WARN] "+format, args...)
}

// Errorf prints logs with ERROR level
func Errorf(format string, args ...interface{}) {
	if !shouldPrint(levelError) {
		return
	}
	log.Printf("[ERROR] "+format, args...)
}

func shouldPrint(l level) bool {
	return l >= logLevel
}
