package workers

// Error is constant error
type Error string

const (
	// ErrMissingKey when config key for worker is missing
	ErrMissingKey = Error("missing worker key")

	// ErrEmptyConfig when viper not passed to params
	ErrEmptyConfig = Error("empty config")

	// ErrEmptyWorkers when workers not passed to params
	ErrEmptyWorkers = Error("empty workers")

	// ErrEmptyJob when worker job is nil
	ErrEmptyJob = Error("empty job")
)

// Error converts constant error from string
func (e Error) Error() string {
	return string(e)
}
