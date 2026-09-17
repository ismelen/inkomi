package usecases

import (
	"bytes"
	"io"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/test/mocks"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5TransactionUC_Process_Normal_ShouldProcess(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()

	mockSource := &mocks.BooksSourceMock{
		DownloadResult: &book.LibgenDownload{
			Stream: io.NopCloser(bytes.NewBufferString("dummy book data")),
			Ext:    "epub",
		},
	}
	mockProvider := &mocks.BooksProviderMock{
		Source:    mockSource,
		HasMirror: true,
	}

	downloadUC := NewDownloadBookUC(mockProvider)
	mockPush := &mocks.MockPushNotifier{}
	mockCloud := &mocks.MockCloudStorage{}

	uc := NewMd5TransactionUC(mockPush, mockCloud, downloadUC)

	md5Path := filepath.Join(tempDir, "some-md5-hash")
	os.WriteFile(md5Path, []byte("dummy md5 source"), 0644)

	file := &convert.TransactionFile{
		Id:      "file1",
		Name:    "test_md5",
		SrcPath: md5Path,
		Size:    0,
	}
	tran := &convert.Transaction{
		Id: "tran1",
	}

	// Act
	result := uc.Process(file, tran, tempDir)

	// Assert
	assert.NotNil(t, result, "Expected result, got nil")
	if result != nil {
		assert.Equal(t, "test_md5", result.Name, "Expected result name test_md5")
		expectedPath := filepath.Join(tempDir, "tran1", "file1", "test_md5.epub")
		assert.Equal(t, expectedPath, result.Path, "Expected path to match")
	}

	_, err := os.Stat(md5Path)
	assert.NoError(t, err, "Expected source file to not be removed by MD5TransactionUC")
}
