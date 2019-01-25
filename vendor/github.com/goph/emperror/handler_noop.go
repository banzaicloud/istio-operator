package emperror

type noopHandler struct{}

// NewNoopHandler creates a no-op error handler that discards all received errors.
// Useful in examples and as a fallback error handler.
func NewNoopHandler() Handler {
	return &noopHandler{}
}

func (*noopHandler) Handle(err error) {}
