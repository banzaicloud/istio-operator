package emperror

// compositeHandler allows an error to be processed by multiple handlers.
type compositeHandler struct {
	handlers []Handler
}

// NewCompositeHandler returns a new compositeHandler.
func NewCompositeHandler(handlers ...Handler) Handler {
	return &compositeHandler{handlers}
}

// Handle goes through the handlers and call each of them for the error.
func (h *compositeHandler) Handle(err error) {
	for _, handler := range h.handlers {
		handler.Handle(err)
	}
}
