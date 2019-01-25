package emperror

// HandleRecover recovers from a panic and handles the error.
//
// 		defer emperror.HandleRecover(errorHandler)
func HandleRecover(handler Handler) {
	err := Recover(recover())
	if err != nil {
		handler.Handle(err)
	}
}

// Panic panics if the passed error is not nil.
//
// This function is useful with HandleRecover when panic is used as a flow control tool to stop the application.
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// Handle handles an error whenever it occurs.
func Handle(handler Handler, err error) {
	if err != nil {
		handler.Handle(err)
	}
}
