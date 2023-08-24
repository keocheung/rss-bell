package logger

import (
	"log"
	"os"
)

type level int

const (
	levelConfigEnvKey       = "LOG_LEVEL"
	levelDebug        level = iota
	levelInfo
	levelWarn
	levelError
)

type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var levelMap = map[string]level{
	"DEBUG": levelDebug,
	"INFO":  levelInfo,
	"WARN":  levelWarn,
	"ERROR": levelError,
}

func Debugf(format string, args ...interface{}) {
	if !shouldPrint(levelDebug) {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func Infof(format string, args ...interface{}) {
	if !shouldPrint(levelInfo) {
		return
	}
	log.Printf("[INFO] "+format, args...)
}

func Warnf(format string, args ...interface{}) {
	if !shouldPrint(levelWarn) {
		return
	}
	log.Printf("[WARN] "+format, args...)
}

func Errorf(format string, args ...interface{}) {
	if !shouldPrint(levelError) {
		return
	}
	log.Printf("[ERROR] "+format, args...)
}

func shouldPrint(l level) bool {
	levelConfigStr := os.Getenv(levelConfigEnvKey)
	levelConfig, ok := levelMap[levelConfigStr]
	if !ok {
		return l >= levelInfo
	}
	return l >= levelConfig
}
