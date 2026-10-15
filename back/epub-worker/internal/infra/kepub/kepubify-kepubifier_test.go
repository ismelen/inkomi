package kepub_test

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/kepub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeEpub(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "book.epub")
	file, err := os.Create(path)
	require.NoError(t, err)
	zw := zip.NewWriter(file)

	files := map[string]string{
		"mimetype": "application/epub+zip",
		"META-INF/container.xml": `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
<metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:identifier id="id">x</dc:identifier><dc:title>Book</dc:title><dc:language>en</dc:language></metadata>
<manifest><item id="c1" href="c1.xhtml" media-type="application/xhtml+xml"/></manifest>
<spine><itemref idref="c1"/></spine>
</package>`,
		"OEBPS/c1.xhtml": `<?xml version="1.0" encoding="utf-8"?>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>c1</title></head><body><p>Hello world. Second sentence.</p></body></html>`,
	}
	for _, name := range []string{"mimetype", "META-INF/container.xml", "OEBPS/content.opf", "OEBPS/c1.xhtml"} {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(files[name]))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	require.NoError(t, file.Close())
	return path
}

func TestKepubifyKepubifier_Kepubify_ValidEpub_AddsKoboSpans(t *testing.T) {
	t.Parallel()

	// Arrange
	out := filepath.Join(t.TempDir(), "book.kepub.epub")

	// Act
	err := kepub.NewKepubifyKepubifier().Kepubify(context.Background(), writeEpub(t), out)

	// Assert
	require.NoError(t, err)
	zr, err := zip.OpenReader(out)
	require.NoError(t, err)
	defer zr.Close()

	rc, err := zr.Open("OEBPS/c1.xhtml")
	require.NoError(t, err)
	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Contains(t, string(content), "koboSpan")
}

func TestKepubifyKepubifier_Kepubify_InvalidEpub_ReturnsErrorAndNoOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	in := filepath.Join(t.TempDir(), "bad.epub")
	require.NoError(t, os.WriteFile(in, []byte("not a zip"), 0o600))
	out := filepath.Join(t.TempDir(), "bad.kepub.epub")

	// Act
	err := kepub.NewKepubifyKepubifier().Kepubify(context.Background(), in, out)

	// Assert
	assert.Error(t, err)
	assert.NoFileExists(t, out)
}
