package cloud_test

import (
	"testing"

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

func TestCloudflareR2Storage_GetUrl_ValidIdAndExt_ReturnsPresignedUrl(t *testing.T) {
	// Arrange
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")
	assert.NoError(t, err)

	id := "document-123"
	ext := ".epub"

	// Act
	url, err := storage.GetUrl(id, ext)
	
	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, url)

	expectedEndpointFragment := "test-account.r2.cloudflarestorage.com"
	assert.Contains(t, url, expectedEndpointFragment)

	expectedPathFragment := "test-bucket.test-account.r2.cloudflarestorage.com/document-123.epub"
	assert.Contains(t, url, expectedPathFragment)
	assert.Contains(t, url, "X-Amz-Signature=")
}

func TestCloudflareR2Storage_Check_InvalidCredentials_ReturnsFalse(t *testing.T) {
	// Arrange
	storage, err := cloud.NewCloudflareR2Storage("test-account", "test-key", "test-secret", "test-bucket")
	assert.NoError(t, err)

	id := "document-123"
	ext := ".epub"

	// Act
	exists := storage.Check(id, ext)

	// Assert
	// Since credentials and account are fake, it should fail to do HeadObject and return false
	assert.False(t, exists)
}
