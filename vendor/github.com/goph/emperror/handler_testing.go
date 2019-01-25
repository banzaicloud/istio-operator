package emperror

import "sync"

// TestHandler is a simple stub for the handler interface recording every error.
//
// The TestHandler is safe for concurrent use.
type TestHandler struct {
	errors []error

	mu sync.RWMutex
}

// NewTestHandler returns a new TestHandler.
func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// Count returns the number of events recorded in the logger.
func (h *TestHandler) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.errors)
}

// LastError returns the last handled error (if any).
func (h *TestHandler) LastError() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.errors) < 1 {
		return nil
	}

	return h.errors[len(h.errors)-1]
}

// Errors returns all handled errors.
func (h *TestHandler) Errors() []error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.errors
}

// Handle records the error.
func (h *TestHandler) Handle(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.errors = append(h.errors, err)
}
