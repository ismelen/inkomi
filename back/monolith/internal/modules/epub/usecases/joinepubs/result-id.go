package joinepubs

import (
	"fmt"

	"github.com/google/uuid"
)

var resultNamespace = uuid.MustParse("6f1c4b2e-8a3d-4c5e-9b7a-2d1e0f3a4b5c")

// ResultId returns the UUID of the index-th result of a parent; retries of the same join get the same id.
func ResultId(parentId string, index int) string {
	return uuid.NewSHA1(resultNamespace, fmt.Appendf(nil, "%s:%d", parentId, index)).String()
}
