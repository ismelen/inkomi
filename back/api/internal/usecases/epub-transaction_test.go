package usecases

import (
	"context"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/test/mocks"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEpubTransactionUC_Process_Normal_ShouldProcess(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()

	mockPush := &mocks.MockPushNotifier{}
	mockCloud := &mocks.MockCloudStorage{}

	uc := NewEpubTransactionUC(mockPush, mockCloud)

	epubPath := filepath.Join(tempDir, "test.epub")
	os.WriteFile(epubPath, []byte("dummy data"), 0644)

	file := &convert.TransactionFile{
		Id:      "file1",
		Name:    "test_epub",
		SrcPath: epubPath,
		Size:    1024,
	}
	tran := &convert.Transaction{
		Id: "tran1",
	}

	// Act
	result := uc.Process(context.Background(), file, tran, tempDir)

	// Assert
	assert.NotNil(t, result, "Expected result, got nil")
	assert.Equal(t, "test_epub", result.Name, "Expected result name test_epub")
	assert.Equal(t, int64(1024), result.Size, "Expected size 1024")
	assert.Len(t, result.Files, 1, "Expected 1 file in result.Files")
	assert.Equal(t, "file1", result.Files[0].Id, "Expected to contain the original file in result.Files")

	_, err := os.Stat(epubPath)
	assert.NoError(t, err, "Expected source file to not be removed by EpubTransactionUC")
}
