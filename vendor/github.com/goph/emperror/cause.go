package emperror

// ForEachCause loops through an error chain and calls a function for each of them,
// starting with the topmost one.
//
// The function can return false to break the loop before it ends.
func ForEachCause(err error, fn func(err error) bool) {
	// causer is the interface defined in github.com/pkg/errors for specifying a parent error.
	type causer interface {
		Cause() error
	}

	for err != nil {
		continueLoop := fn(err)
		if !continueLoop {
			break
		}

		cause, ok := err.(causer)
		if !ok {
			break
		}

		err = cause.Cause()
	}
}
