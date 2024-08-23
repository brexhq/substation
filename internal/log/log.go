// Package log wraps logrus and provides global logging
// only debug logging should be used in condition/, process/, and internal/ to reduce the likelihood of corrupting output for apps
// debug and info logging can be used in cmd/
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Debug wraps logrus Debug function with stack information
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info wraps logrus Debug function with stack information
func Info(args ...interface{}) {
	log.Info(args...)
}

// WithField wraps logrus WithField function
func WithField(k string, v interface{}) *logrus.Entry {
	return log.WithField(k, v)
}

func init() {
	if _, ok := os.LookupEnv("SUBSTATION_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
		return
	}

	log.SetLevel(logrus.InfoLevel)
}
