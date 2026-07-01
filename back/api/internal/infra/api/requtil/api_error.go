package requtil

type ApiError struct {
	Status  int
	Message string
}

func New(status int, message string) *ApiError {
	return &ApiError{Status: status, Message: message}
}

func (e *ApiError) Error() string {
	return e.Message
}
