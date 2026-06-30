package apierr

// ApiError represents an HTTP-level error with a status code and message.
// This type lives in infra/api because HTTP status codes are infrastructure concerns.
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
