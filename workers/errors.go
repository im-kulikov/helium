package workers

type Error string

const (
	ErrMissingKey     = Error("missing worker key")
	ErrRedisClientNil = Error("gotten nil redis client for exclusive worker")
)

func (e Error) Error() string {
	return string(e)
}
