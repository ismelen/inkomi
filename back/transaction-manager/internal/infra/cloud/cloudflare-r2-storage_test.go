package cloud_test

import (
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/cloud"
	"github.com/stretchr/testify/assert"
)

func TestCloudflareR2Storage_New_ValidConfig_ReturnsInstance(t *testing.T) {
	// Arrange & Act
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestCloudflareR2Storage_GetUrl_ValidIdAndFilename_ReturnsPresignedUrl(t *testing.T) {
	// Arrange
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")
	assert.NoError(t, err)

	id := "document-123"
	filename := "book.epub"

	// Act
	url, err := storage.GetUrl(id, filename)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, url)

	expectedEndpointFragment := "test-account.r2.cloudflarestorage.com"
	assert.Contains(t, url, expectedEndpointFragment)

	expectedPathFragment := "test-bucket.test-account.r2.cloudflarestorage.com/document-123/book.epub"
	assert.Contains(t, url, expectedPathFragment)
	assert.Contains(t, url, "X-Amz-Signature=")
}

func TestCloudflareR2Storage_Check_InvalidCredentials_ReturnsFalse(t *testing.T) {
	// Arrange
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")
	assert.NoError(t, err)

	id := "document-123"
	filename := "book.epub"

	// Act
	exists := storage.Check(id, filename)

	// Assert
	// Since credentials and account are fake, it should fail to do HeadObject and return false
	assert.False(t, exists)
}

func TestCloudflareR2Storage_GetDownloadUrl_ValidIdAndFilename_ReturnsAttachmentUrl(t *testing.T) {
	// Arrange
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")
	assert.NoError(t, err)

	// Act
	url, expiresAt, err := storage.GetDownloadUrl("document-123", "Mi libro ñ.epub")

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, url, "test-bucket.test-account.r2.cloudflarestorage.com/document-123/")
	assert.Contains(t, url, "response-content-disposition=attachment")
	assert.Contains(t, url, "X-Amz-Signature=")
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), expiresAt, time.Minute)
}
