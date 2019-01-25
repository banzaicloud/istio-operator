// Package logrushandler provides Logrus integration.
package logrushandler

import (
	"github.com/goph/emperror"
	"github.com/goph/emperror/internal/keyvals"
	"github.com/sirupsen/logrus"
)

// Handler logs errors using Logrus.
type Handler struct {
	logger logrus.FieldLogger
}

// New creates a new handler.
func New(logger logrus.FieldLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// Handle logs an error.
func (h *Handler) Handle(err error) {
	logger := h.logger

	// Extract context from the error and attach it to the log
	if kvs := emperror.Context(err); len(kvs) > 0 {
		logger = h.logger.WithFields(logrus.Fields(keyvals.ToMap(kvs)))
	}

	type errorCollection interface {
		Errors() []error
	}

	if errs, ok := err.(errorCollection); ok {
		for _, e := range errs.Errors() {
			logger.WithField("parent", err.Error()).Error(e.Error())
		}
	} else {
		logger.Error(err.Error())
	}
}
