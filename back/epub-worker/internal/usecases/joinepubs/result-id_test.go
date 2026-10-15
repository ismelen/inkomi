package joinepubs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResultId_SameInput_ReturnsSameUUID(t *testing.T) {
	t.Parallel()

	// Act
	first := joinepubs.ResultId("parent-1", 0)
	again := joinepubs.ResultId("parent-1", 0)
	other := joinepubs.ResultId("parent-1", 1)

	// Assert
	_, err := uuid.Parse(first)
	require.NoError(t, err)
	assert.Equal(t, first, again)
	assert.NotEqual(t, first, other)
}
