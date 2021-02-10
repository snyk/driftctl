package errors

type UsageError struct {
	msg string
}

func NewUsageError(msg string) UsageError {
	return UsageError{msg}
}

func (u UsageError) Error() string {
	return u.msg
}
