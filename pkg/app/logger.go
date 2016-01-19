package app

import "github.com/Sirupsen/logrus"

// DefaultLogger builds an instance of the default logger.
func DefaultLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{})
}
