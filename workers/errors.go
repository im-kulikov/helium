package workers

import "github.com/im-kulikov/helium/internal"

const (
	// ErrMissingKey when config key for worker is missing
	ErrMissingKey = internal.Error("missing worker key")

	// ErrEmptyConfig when viper not passed to params
	ErrEmptyConfig = internal.Error("empty config")

	// ErrEmptyWorkers when workers not passed to params
	ErrEmptyWorkers = internal.Error("empty workers")

	// ErrEmptyLocker when locker required,
	// but not passed to params
	ErrEmptyLocker = internal.Error("empty locker")

	// ErrEmptyJob when worker job is nil
	ErrEmptyJob = internal.Error("empty job")
)
