package usecases

import (
	"archive/zip"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/test/mocks"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestCBZ(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Create 2 test images
	for i := 1; i <= 2; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 800, 1200))
		for y := 0; y < 1200; y++ {
			for x := 0; x < 800; x++ {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}

		fWriter, err := w.Create(string(rune('0'+i)) + ".jpg")
		if err != nil {
			t.Fatal(err)
		}
		if err := jpeg.Encode(fWriter, img, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMangaTransactionUC_Process_Normal_ShouldProcess(t *testing.T) {
	t.Parallel()
	// Arrange
	tempDir := t.TempDir()

	cbzPath := filepath.Join(tempDir, "test_manga.cbz")
	createTestCBZ(t, cbzPath)

	mockPush := &mocks.MockPushNotifier{}
	mockCloud := &mocks.MockCloudStorage{}

	uc := NewMangaTransactionUC(mockPush, mockCloud)

	file := &convert.TransactionFile{
		Id:      "file1",
		Name:    "test_manga",
		SrcPath: cbzPath,
		Size:    1000,
	}
	tran := &convert.Transaction{
		Id: "tran1",
		Config: &convert.TransactionConfig{
			Profile: &convert.EReaderProfile{
				Width:  800,
				Height: 1200,
			},
		},
	}

	// Act
	result := uc.Process(context.Background(), file, tran, tempDir)

	// Assert
	assert.NotNil(t, result, "Expected result, got nil. Error: %v", file.Error)
	if result != nil {
		assert.Equal(t, "test_manga.epub", result.Filename, "Expected result name test_manga.epub")
		_, err := os.Stat(result.Path)
		assert.False(t, os.IsNotExist(err), "Expected output epub to exist at %s", result.Path)
	}

	_, err := os.Stat(cbzPath)
	assert.True(t, os.IsNotExist(err), "Expected source file to be removed by MangaTransactionUC")
}
