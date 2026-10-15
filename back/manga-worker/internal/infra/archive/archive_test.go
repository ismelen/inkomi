package archive_test

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/archive"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeZip(t *testing.T, entries map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "chapter.cbz")
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	w := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := w.Create(name)
		require.NoError(t, err)
		_, err = io.WriteString(entry, content)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return path
}

func jpegBytes(t *testing.T, w, h int) []byte {
	t.Helper()

	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = color.Gray{Y: uint8(i)}.Y
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, nil))
	return buf.Bytes()
}

func TestCBZExtractor_Extract_MixedEntries_ReturnsImagesInNaturalOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	src := writeZip(t, map[string]string{
		"vol/10.jpg":           "ten",
		"vol/2.PNG":            "two",
		"vol/1.jpeg":           "one",
		"vol/notes.txt":        "skip",
		"__MACOSX/vol/._1.jpg": "skip",
		"vol/.hidden.jpg":      "skip",
		"../../escape.jpg":     "zero",
	})
	dst := t.TempDir()

	// Act
	paths, err := archive.NewCBZExtractor().Extract(context.Background(), src, dst)

	// Assert
	require.NoError(t, err)
	require.Len(t, paths, 4)
	want := []struct{ name, content string }{
		{"0001.jpg", "zero"},
		{"0002.jpeg", "one"},
		{"0003.png", "two"},
		{"0004.jpg", "ten"},
	}
	for i, w := range want {
		assert.Equal(t, filepath.Join(dst, w.name), paths[i])
		content, err := os.ReadFile(paths[i])
		require.NoError(t, err)
		assert.Equal(t, w.content, string(content))
	}
}

func TestCBZExtractor_Extract_NotAZip_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	src := filepath.Join(t.TempDir(), "broken.cbz")
	require.NoError(t, os.WriteFile(src, []byte("not a zip"), 0o644))

	// Act
	_, err := archive.NewCBZExtractor().Extract(context.Background(), src, t.TempDir())

	// Assert
	assert.Error(t, err)
}

func TestPDFExtractor_Extract_ImagePages_ReturnsOneImagePerPage(t *testing.T) {
	t.Parallel()

	// Arrange
	src := filepath.Join(t.TempDir(), "chapter.pdf")
	out, err := os.Create(src)
	require.NoError(t, err)
	imgs := []io.Reader{bytes.NewReader(jpegBytes(t, 40, 60)), bytes.NewReader(jpegBytes(t, 50, 70))}
	require.NoError(t, api.ImportImages(nil, out, imgs, nil, nil))
	require.NoError(t, out.Close())
	dst := t.TempDir()

	// Act
	paths, err := archive.NewPDFExtractor().Extract(context.Background(), src, dst)

	// Assert
	require.NoError(t, err)
	require.Len(t, paths, 2)
	for i, path := range paths {
		file, err := os.Open(path)
		require.NoError(t, err)
		cfg, _, err := image.DecodeConfig(file)
		require.NoError(t, file.Close())
		require.NoError(t, err, "page %d", i)
		assert.Equal(t, []int{40, 50}[i], cfg.Width)
	}
}
