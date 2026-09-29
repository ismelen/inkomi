package usecases

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/test/mocks"

	"github.com/stretchr/testify/assert"
)

func TestBaseTransactionUC_Execute_Normal_ShouldProcess(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "test.cbz")
	os.WriteFile(srcFile, []byte("dummy data"), 0644)

	resFile := filepath.Join(tempDir, "out.epub")
	os.WriteFile(resFile, []byte("dummy result data"), 0644)

	mockPush := &mocks.MockPushNotifier{}
	mockCloud := &mocks.MockCloudStorage{}
	mockProcessor := &mocks.MockTransactionUC{
		ProcessResult: &convert.TransactionResultFile{
			Name: "out.epub",
			Path: resFile,
			Size: 100,
		},
	}

	baseUC := BaseTransactionUC{
		pushNotifier: mockPush,
		cloud:        mockCloud,
		processor:    mockProcessor,
	}

	file := &convert.TransactionFile{
		Id:      "file1",
		SrcPath: srcFile,
	}
	config := &convert.TransactionConfig{
		Cloud: false,
		Merge: false,
		Profile: &convert.EReaderProfile{
			IsKepub: false,
		},
	}
	tran := convert.NewTransaction("tran1", config, tempDir)

	// Act
	baseUC.Execute(context.Background(), file, tran, tempDir)

	// Assert
	assert.True(t, mockPush.SendCalled, "Expected push notification to be sent")

	assert.Equal(t, convert.TransactionFileDone, file.Status(), "Expected file status to be Done")
	assert.Equal(t, convert.TransactionDone, tran.Status.Get(), "Expected transaction status to be Done")
}

func TestBaseTransactionUC_Execute_SendAndNotify_ShouldNotify(t *testing.T) {
	t.Parallel()
	// Arrange
	mockPush := &mocks.MockPushNotifier{}
	mockCloud := &mocks.MockCloudStorage{}

	baseUC := BaseTransactionUC{
		pushNotifier: mockPush,
		cloud:        mockCloud,
	}

	tran := &convert.Transaction{
		Id: "tran1",
		Config: &convert.TransactionConfig{
			Cloud:       true,
			NotifyToken: "token123",
			CloudToken:  "cloud_token",
			CloudFolder: "folder1",
		},
		OnFreeSpace: func(size int32) {},
	}
	result := &convert.TransactionResultFile{
		Name: "test.epub",
		Path: "/fake/path/test.epub",
	}

	// Act
	baseUC.SendAndNotify(tran, result)

	// Assert
	assert.True(t, mockPush.SendCalled, "Expected push notification to be sent")
	assert.True(t, mockCloud.UploadCalled, "Expected cloud upload to be called")
	assert.Equal(t, "/fake/path/test.epub", mockCloud.LastPath, "Expected cloud upload path to match")
}
