package emperror

import (
	"fmt"

	"github.com/pkg/errors"
)

// Recover accepts a recovered panic (if any) and converts it to an error (if necessary).
func Recover(r interface{}) (err error) {
	if r != nil {
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			if _, ok := x.(stackTracer); !ok {
				x = errors.WithStack(x)
			}

			err = x
		default:
			err = errors.New(fmt.Sprintf("unknown panic, received: %v", r))
		}
	}

	return err
}
