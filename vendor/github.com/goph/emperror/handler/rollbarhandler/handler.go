// Package rollbarhandler provides Rollbar integration.
package rollbarhandler

import (
	"github.com/goph/emperror"
	"github.com/goph/emperror/httperr"
	"github.com/goph/emperror/internal/keyvals"
	"github.com/rollbar/rollbar-go"
)

// Handler is responsible for sending errors to Rollbar.
type Handler struct {
	client *rollbar.Client
}

// New creates a new handler.
func New(token, environment, codeVersion, serverHost, serverRoot string) *Handler {
	return NewFromClient(rollbar.New(token, environment, codeVersion, serverHost, serverRoot))
}

// NewFromClient creates a new handler from a client instance.
func NewFromClient(client *rollbar.Client) *Handler {
	return &Handler{
		client: client,
	}
}

// Handle sends the error to Rollbar.
func (h *Handler) Handle(err error) {
	// Get the context from the error
	ctx := keyvals.ToMap(emperror.Context(err))

	// Expose the stackTracer interface on the outer error (if there is stack trace in the error)
	// Convert error with stack trace to an internal error type
	if e, ok := emperror.ExposeStackTrace(err).(stackTracer); ok {
		err = newCauseStacker(e)
	}

	if req, ok := httperr.HTTPRequest(err); ok {
		h.client.RequestErrorWithStackSkipWithExtras(rollbar.ERR, req, err, 3, ctx)

		return
	}

	h.client.ErrorWithStackSkipWithExtras(rollbar.ERR, err, 3, ctx)
}

// Close closes the underlying notifier and waits for asynchronous reports to finish.
func (h *Handler) Close() error {
	return h.client.Close()
}
