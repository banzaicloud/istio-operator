package emperror

import (
	"fmt"
	"io"
	"runtime"

	"github.com/pkg/errors"
)

func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr

	n := runtime.Callers(3, pcs[:])

	return pcs[0:n]
}

type wrappedError struct {
	err   error
	stack []uintptr
}

func (e *wrappedError) Error() string {
	return e.err.Error()
}

func (e *wrappedError) Cause() error {
	return e.err
}

func (e *wrappedError) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(e.stack))

	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((e.stack)[i])
	}

	return f
}

func (e *wrappedError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", e.Cause())

			for _, pc := range e.stack {
				f := errors.Frame(pc)
				_, _ = fmt.Fprintf(s, "\n%+v", f)
			}

			return
		}

		fallthrough

	case 's':
		_, _ = io.WriteString(s, e.Error())

	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called (if there is none attached to the error yet), and the supplied message.
// If err is nil, Wrap returns nil.
//
// Note: do not use this method when passing errors between goroutines.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	_, ok := getStackTracer(err)

	err = errors.WithMessage(err, message)

	// There is no stack trace in the error, so attach it here
	if !ok {
		err = &wrappedError{
			err:   err,
			stack: callers(),
		}
	}

	return err
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is call (if there is none attached to the error yet), and the format specifier.
// If err is nil, Wrapf returns nil.
//
// Note: do not use this method when passing errors between goroutines.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	_, ok := getStackTracer(err)

	err = errors.WithMessage(err, fmt.Sprintf(format, args...))

	// There is no stack trace in the error, so attach it here
	if !ok {
		err = &wrappedError{
			err:   err,
			stack: callers(),
		}
	}

	return err
}
