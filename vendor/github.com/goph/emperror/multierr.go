package emperror

// multiError aggregates multiple errors into a single value.
//
// While ErrorCollection is only an interface for listing errors,
// multiError actually implements the error interface so it can be returned as an error.
type multiError struct {
	errors []error
	msg    string
}

// Error implements the error interface.
func (e *multiError) Error() string {
	if e.msg != "" {
		return e.msg
	}

	return "Multiple errors happened"
}

// Errors returns the list of wrapped errors.
func (e *multiError) Errors() []error {
	return e.errors
}

// SingleWrapMode defines how MultiErrorBuilder behaves when there is only one error in the list.
type SingleWrapMode int

// These constants cause MultiErrorBuilder to behave as described if there is only one error in the list.
const (
	AlwaysWrap   SingleWrapMode = iota // Always return a MultiError.
	ReturnSingle                       // Return the single error.
)

// MultiErrorBuilder provides an interface for aggregating errors and exposing them as a single value.
type MultiErrorBuilder struct {
	errors []error

	Message        string
	SingleWrapMode SingleWrapMode
}

// NewMultiErrorBuilder returns a new MultiErrorBuilder.
func NewMultiErrorBuilder() *MultiErrorBuilder {
	return &MultiErrorBuilder{
		SingleWrapMode: AlwaysWrap,
	}
}

// Add adds an error to the list.
//
// Calling this method concurrently is not safe.
func (b *MultiErrorBuilder) Add(err error) {
	// Do not add nil values.
	if err == nil {
		return
	}

	b.errors = append(b.errors, err)
}

// ErrOrNil returns a MultiError the builder aggregates a list of errors,
// or returns nil if the list of errors is empty.
//
// It is useful to avoid checking if there are any errors added to the list.
func (b *MultiErrorBuilder) ErrOrNil() error {
	// No errors added, return nil.
	if len(b.errors) == 0 {
		return nil
	}

	// Return a single error when there is only one and the builder is told to do so.
	if len(b.errors) == 1 && b.SingleWrapMode == ReturnSingle {
		return b.errors[0]
	}

	return &multiError{b.errors, b.Message}
}
