package emperror

// The implementation bellow is heavily influenced by go-kit's log context.

// HandlerWith returns a new error handler with keyvals context appended to it.
// If the wrapped error handler is already a contextual error handler created by HandlerWith or HandlerWithPrefix
// keyvals is appended to the existing context, but a new error handler is returned.
//
// The created handler will prepend it's own context to the handled errors.
func HandlerWith(handler Handler, keyvals ...interface{}) Handler {
	if len(keyvals) == 0 {
		return handler
	}

	kvs, handler := extractHandlerContext(handler)

	kvs = append(kvs, keyvals...)

	if len(kvs)%2 != 0 {
		kvs = append(kvs, nil)
	}

	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in HandlerWith.
	// Using the extra capacity without copying risks a data race.
	return newContextualHandler(handler, kvs[:len(kvs):len(kvs)])
}

// HandlerWithPrefix returns a new error handler with keyvals context prepended to it.
// If the wrapped error handler is already a contextual error handler created by HandlerWith or HandlerWithPrefix
// keyvals is prepended to the existing context, but a new error handler is returned.
//
// The created handler will prepend it's own context to the handled errors.
func HandlerWithPrefix(handler Handler, keyvals ...interface{}) Handler {
	if len(keyvals) == 0 {
		return handler
	}

	prevkvs, handler := extractHandlerContext(handler)

	n := len(prevkvs) + len(keyvals)
	if len(keyvals)%2 != 0 {
		n++
	}

	kvs := make([]interface{}, 0, n)
	kvs = append(kvs, keyvals...)

	if len(kvs)%2 != 0 {
		kvs = append(kvs, nil)
	}

	kvs = append(kvs, prevkvs...)

	return newContextualHandler(handler, kvs)
}

// extractHandlerContext extracts the context and optionally the wrapped handler when it's the same container.
func extractHandlerContext(handler Handler) ([]interface{}, Handler) {
	var kvs []interface{}

	if c, ok := handler.(*contextualHandler); ok {
		handler = c.handler
		kvs = c.keyvals
	}

	return kvs, handler
}

// contextualHandler is a Handler implementation returned by HandlerWith or HandlerWithPrefix.
//
// It wraps an error handler and a holds keyvals as the context.
type contextualHandler struct {
	handler Handler
	keyvals []interface{}
}

// newContextualHandler creates a new *contextualHandler or a struct which is contextual and holds a stack trace.
func newContextualHandler(handler Handler, kvs []interface{}) Handler {
	chandler := &contextualHandler{
		handler: handler,
		keyvals: kvs,
	}

	return chandler
}

// Handle prepends the handler's context to the error's (if any) and delegates the call to the underlying handler.
func (h *contextualHandler) Handle(err error) {
	err = With(err, h.keyvals...)

	h.handler.Handle(err)
}
