package domain

type ApiError struct {
	Status  int
	Message string
}

func NewApiError(status int, message string) ApiError {
	return ApiError{
		Status:  status,
		Message: message,
	}
}

func (e ApiError) Error() string {
	return e.Message
}
