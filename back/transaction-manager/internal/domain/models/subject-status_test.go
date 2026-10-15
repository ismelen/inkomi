package models_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestStatusFromSubject_Joined_ReturnsJoined(t *testing.T) {
	t.Parallel()

	// Act
	status := models.StatusFromSubject(jobs.SubjectJoined)

	// Assert
	assert.Equal(t, models.SourceStatusJoined, status)
}
