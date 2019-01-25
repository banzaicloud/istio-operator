package sentryhandler

import (
	"github.com/getsentry/raven-go"
	"github.com/goph/emperror"
	"github.com/goph/emperror/httperr"
	"github.com/goph/emperror/internal/keyvals"
	"github.com/pkg/errors"
)

// Handler is responsible for sending errors to Sentry.
type Handler struct {
	client *raven.Client

	sendSynchronously bool
}

// New creates a new handler.
func New(dsn string) (*Handler, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raven client")
	}

	return NewFromClient(client), nil
}

// NewSync creates a new handler that sends errors synchronously.
func NewSync(dsn string) (*Handler, error) {
	handler, err := New(dsn)
	if err != nil {
		return nil, err
	}

	handler.sendSynchronously = true

	return handler, nil
}

// NewFromClient creates a new handler from a client instance.
func NewFromClient(client *raven.Client) *Handler {
	return &Handler{
		client: client,
	}
}

// NewSyncFromClient creates a new handler from a client instance that sends errors synchronously.
func NewSyncFromClient(client *raven.Client) *Handler {
	handler := NewFromClient(client)

	handler.sendSynchronously = true

	return handler
}

// Handle sends the error to Rollbar.
func (h *Handler) Handle(err error) {
	var interfaces []raven.Interface

	// Get HTTP request (if any)
	if req, ok := httperr.HTTPRequest(err); ok {
		interfaces = append(interfaces, raven.NewHttp(req))
	}

	packet := raven.NewPacketWithExtra(
		err.Error(),
		keyvals.ToMap(emperror.Context(err)),
		append(
			interfaces,
			raven.NewException(
				err,
				raven.GetOrNewStacktrace(emperror.ExposeStackTrace(err), 1, 3, h.client.IncludePaths()),
			),
		)...,
	)

	eventID, ch := h.client.Capture(packet, nil)

	if h.sendSynchronously && eventID != "" {
		<-ch
	}
}

// Close closes the underlying notifier and waits for asynchronous reports to finish.
func (h *Handler) Close() error {
	h.client.Close()
	h.client.Wait()

	return nil
}
