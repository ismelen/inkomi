package requtil

type ApiError struct {
	Status  int
	Message string
}

func (e *ApiError) Error() string { return e.Message }

func NewError(status int, message string) *ApiError {
	return &ApiError{Status: status, Message: message}
}
