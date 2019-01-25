package bugsnaghandler

import (
	berrors "github.com/bugsnag/bugsnag-go/errors"
	"github.com/pkg/errors"
)

type stackTracer interface {
	Error() string
	StackTrace() errors.StackTrace
}

type errorWithStackFrames struct {
	err stackTracer
}

// newErrorWithStackFrames returns a new error implementing the
// github.com/bugsnag/bugsnag-go/errors.ErrorWithStackFrames interface.
func newErrorWithStackFrames(err stackTracer) *errorWithStackFrames {
	return &errorWithStackFrames{err}
}

// Error implements the error interface.
func (e *errorWithStackFrames) Error() string {
	return e.err.Error()
}

// Cause implements the github.com/pkg/errors.causer interface.
func (e *errorWithStackFrames) Cause() error {
	return e.err
}

// StackTrace implements the github.com/pkg/errors.stackTracer interface.
func (e *errorWithStackFrames) StackTrace() errors.StackTrace {
	return e.err.StackTrace()
}

func (e *errorWithStackFrames) StackFrames() []berrors.StackFrame {
	stackTrace := e.err.StackTrace()
	stackFrames := make([]berrors.StackFrame, len(stackTrace))

	for i, frame := range stackTrace {
		stackFrames[i] = berrors.NewStackFrame(uintptr(frame))
	}

	return stackFrames
}
