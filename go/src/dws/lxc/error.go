package dws_lxc

var (
	ErrNoTemplatesInstalled = NewError("lxc-templates not installed")
	ErrNoTemplatesFound     = NewError("no templates found")
)

// Error represents a basic error that implies the error interface.
type Error struct {
	Message string
}

// NewError creates a new error with the given msg argument.
func NewError(msg string) error {
	return &Error{
		Message: msg,
	}
}

func (e *Error) Error() string {
	return e.Message
}
