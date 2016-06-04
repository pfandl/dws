package err

import "strings"

// Error represents a basic error that implies the error interface.
type _error struct {
	Message string
}

// Error creates a new error with the given msg argument.
func New(s ...string) error {
	return &_error{
		Message: strings.Join(s, " | "),
	}
}

func (e *_error) Error() string {
	return e.Message
}
