package rollbarhandler

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rollbar/rollbar-go"
)

type stackTracer interface {
	Error() string
	StackTrace() errors.StackTrace
}

type causeStacker struct {
	err stackTracer
}

func newCauseStacker(err stackTracer) *causeStacker {
	return &causeStacker{
		err: err,
	}
}

func (e *causeStacker) Error() string {
	return e.err.Error()
}

func (e *causeStacker) Cause() error {
	if c, ok := e.err.(interface{ Cause() error }); ok {
		return c.Cause()
	}

	return nil
}

func (e *causeStacker) Stack() rollbar.Stack {
	stackTrace := e.err.StackTrace()
	stack := make(rollbar.Stack, len(stackTrace))

	for i, frame := range stackTrace {
		line, _ := strconv.Atoi(fmt.Sprintf("%d", frame))

		stack[i] = rollbar.Frame{
			Filename: shortenFilePath(strings.SplitN(fmt.Sprintf("%+s", frame), "\n\t", 2)[1]), // nolint: govet
			Method:   fmt.Sprintf("%n", frame),                                                 // nolint: govet
			Line:     line,
		}
	}

	return stack
}

// copied from rollbar

// nolint: gochecknoglobals
var (
	knownFilePathPatterns = []string{
		"github.com/",
		"code.google.com/",
		"bitbucket.org/",
		"launchpad.net/",
	}
)

func shortenFilePath(s string) string {
	// added to the original function
	s = strings.TrimPrefix(s, runtime.GOROOT())

	idx := strings.Index(s, "/src/pkg/")
	if idx != -1 {
		return s[idx+5:]
	}
	for _, pattern := range knownFilePathPatterns {
		idx = strings.Index(s, pattern)
		if idx != -1 {
			return s[idx:]
		}
	}
	return s
}
