package emperror

import (
	"fmt"
	"io"
)

// The implementation bellow is heavily influenced by go-kit's log context.

// With returns a new error with keyvals context appended to it.
// If the wrapped error is already a contextual error created by With
// keyvals is appended to the existing context, but a new error is returned.
func With(err error, keyvals ...interface{}) error {
	if err == nil {
		return nil
	}

	if len(keyvals) == 0 {
		return err
	}

	var kvs []interface{}

	// extract context from previous error
	if c, ok := err.(*withContext); ok {
		err = c.err

		kvs = append(kvs, c.keyvals...)

		if len(kvs)%2 != 0 {
			kvs = append(kvs, nil)
		}
	}

	kvs = append(kvs, keyvals...)

	if len(kvs)%2 != 0 {
		kvs = append(kvs, nil)
	}

	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in With.
	// Using the extra capacity without copying risks a data race.
	return &withContext{
		err:     err,
		keyvals: kvs[:len(kvs):len(kvs)],
	}
}

// Context extracts the context key-value pairs from an error (or error chain).
func Context(err error) []interface{} {
	type contextor interface {
		Context() []interface{}
	}

	var kvs []interface{}

	ForEachCause(err, func(err error) bool {
		if cerr, ok := err.(contextor); ok {
			kvs = append(cerr.Context(), kvs...)
		}

		return true
	})

	return kvs
}

// withContext annotates an error with context.
type withContext struct {
	err     error
	keyvals []interface{}
}

func (w *withContext) Error() string {
	return w.err.Error()
}

// Context returns the appended keyvals.
func (w *withContext) Context() []interface{} {
	return w.keyvals
}

func (w *withContext) Cause() error {
	return w.err
}

func (w *withContext) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", w.Cause())
			return
		}

		fallthrough

	case 's':
		_, _ = io.WriteString(s, w.Error())

	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.Error())
	}
}
