package requtil

import (
	"encoding/json"
	"net/http"
)

const maxBodySize = 64 << 10

// DecodeJSON parses the JSON body of r into a T, returning a 400 *ApiError when it can't.
func DecodeJSON[T any](r *http.Request) (*T, error) {
	var body T
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodySize))
	if err := decoder.Decode(&body); err != nil {
		return nil, NewError(http.StatusBadRequest, "invalid json body")
	}
	return &body, nil
}
