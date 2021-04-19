package internal

// Error constant type error.
type Error string

// Error returns error message as string.
func (e Error) Error() string { return string(e) }
