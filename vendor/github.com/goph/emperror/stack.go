package emperror

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StackTrace returns the stack trace from an error (if any).
func StackTrace(err error) (errors.StackTrace, bool) {
	st, ok := getStackTracer(err)
	if ok {
		return st.StackTrace(), true
	}

	return nil, false
}

// getStackTracer returns the stack trace from an error (if any).
func getStackTracer(err error) (stackTracer, bool) {
	var st stackTracer

	ForEachCause(err, func(err error) bool {
		if s, ok := err.(stackTracer); ok {
			st = s

			return false
		}

		return true
	})

	return st, st != nil
}

type withStack struct {
	err error
	st  stackTracer
}

func (w *withStack) Error() string {
	return w.err.Error()
}

func (w *withStack) Cause() error {
	return w.err
}

func (w *withStack) StackTrace() errors.StackTrace {
	return w.st.StackTrace()
}

// Format implements the fmt.Formatter interface.
func (w *withStack) Format(s fmt.State, verb rune) {
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

// ExposeStackTrace exposes the stack trace (if any) in the outer error.
func ExposeStackTrace(err error) error {
	if err == nil {
		return err
	}

	st, ok := getStackTracer(err)
	if !ok {
		return err
	}

	return &withStack{
		err: err,
		st:  st,
	}
}
