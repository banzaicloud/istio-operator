package emperror

// Handler is responsible for handling an error.
//
// This interface allows libraries to decouple from logging and error handling solutions.
type Handler interface {
	// Handle takes care of unhandled errors.
	Handle(err error)
}
