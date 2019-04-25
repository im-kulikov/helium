package workers

// Error is constant error
type Error string

const (
	// ErrMissingKey when config key for worker is missing
	ErrMissingKey = Error("missing worker key")

	// ErrRedisClientNil when redis client not provided
	ErrRedisClientNil = Error("gotten nil redis client for exclusive worker")
)

// Error converts constant error from string
func (e Error) Error() string {
	return string(e)
}
