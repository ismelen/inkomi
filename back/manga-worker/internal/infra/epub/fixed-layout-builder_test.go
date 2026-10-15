package epub_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/epub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pageFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func readEntries(t *testing.T, path string) (map[string]string, []string) {
	t.Helper()

	reader, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer reader.Close()

	entries := map[string]string{}
	var order []string
	for _, file := range reader.File {
		rc, err := file.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, rc.Close())
		require.NoError(t, err)
		entries[file.Name] = string(content)
		order = append(order, file.Name)
	}
	return entries, order
}

func buildBook(t *testing.T, rtl bool) (map[string]string, []string) {
	t.Helper()

	dir := t.TempDir()
	pages := []models.EpubPage{
		{Path: pageFile(t, dir, "a.png", "cover"), Width: 100, Height: 150, Format: models.ImageFormatPNG},
		{Path: pageFile(t, dir, "b.jpg", "right"), Width: 50, Height: 150, Split: models.SplitToRight, Format: models.ImageFormatJPEG},
		{Path: pageFile(t, dir, "c.jpg", "left"), Width: 50, Height: 150, Split: models.SplitToLeft, Format: models.ImageFormatJPEG},
	}
	out := filepath.Join(dir, "book.epub")

	err := epub.NewFixedLayoutBuilder().Build(out, models.EpubMeta{Title: "Tom & Jerry", RTL: rtl, Width: 100, Height: 200}, pages)

	require.NoError(t, err)
	return readEntries(t, out)
}

func TestFixedLayoutBuilder_Build_Pages_WritesEpubStructure(t *testing.T) {
	t.Parallel()

	// Act
	entries, order := buildBook(t, false)

	// Assert
	assert.Equal(t, "mimetype", order[0])
	assert.Equal(t, "application/epub+zip", entries["mimetype"])
	for _, name := range []string{
		"META-INF/container.xml", "OEBPS/content.opf", "OEBPS/toc.ncx", "OEBPS/nav.xhtml", "OEBPS/Text/style.css",
		"OEBPS/Images/cover.png", "OEBPS/Images/p0001.png", "OEBPS/Images/p0002.jpg", "OEBPS/Images/p0003.jpg",
		"OEBPS/Text/p0001.xhtml", "OEBPS/Text/p0002.xhtml", "OEBPS/Text/p0003.xhtml",
	} {
		assert.Contains(t, entries, name)
	}
	assert.Equal(t, "cover", entries["OEBPS/Images/cover.png"])
	assert.Equal(t, "left", entries["OEBPS/Images/p0003.jpg"])

	opf := entries["OEBPS/content.opf"]
	assert.Contains(t, opf, `<dc:title>Tom &amp; Jerry</dc:title>`)
	assert.Contains(t, opf, `href="Images/cover.png" media-type="image/png" properties="cover-image"`)
	assert.Contains(t, opf, `<item id="img_p0002" href="Images/p0002.jpg" media-type="image/jpeg"/>`)
	assert.Contains(t, opf, `<meta name="original-resolution" content="100x200"/>`)
	assert.Contains(t, entries["OEBPS/Text/p0002.xhtml"], `src="../Images/p0002.jpg"`)
	assert.Contains(t, entries["OEBPS/Text/p0001.xhtml"], `top:12.5%;`)
}

func TestFixedLayoutBuilder_Build_ReadingDirection_SetsSpineAndSpreads(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		rtl         bool
		progression string
		mode        string
		firstSide   string
	}{
		{"ltr", false, "ltr", "horizontal-lr", "left"},
		{"rtl", true, "rtl", "horizontal-rl", "right"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Act
			entries, _ := buildBook(t, c.rtl)

			// Assert
			opf := entries["OEBPS/content.opf"]
			assert.Contains(t, opf, `page-progression-direction="`+c.progression+`"`)
			assert.Contains(t, opf, `content="`+c.mode+`"`)
			spine := opf[strings.Index(opf, "<spine"):]
			assert.Contains(t, spine, `idref="page_p0001" properties="rendition:page-spread-`+c.firstSide+`"`)
			assert.Contains(t, spine, `idref="page_p0002" properties="rendition:page-spread-right"`)
			assert.Contains(t, spine, `idref="page_p0003" properties="rendition:page-spread-left"`)
		})
	}
}

func TestFixedLayoutBuilder_Build_NoPages_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	err := epub.NewFixedLayoutBuilder().Build(filepath.Join(t.TempDir(), "book.epub"), models.EpubMeta{Title: "x"}, nil)

	// Assert
	assert.Error(t, err)
}

func TestFixedLayoutBuilder_Build_MissingPageFile_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	pages := []models.EpubPage{{Path: filepath.Join(t.TempDir(), "missing.jpg"), Format: models.ImageFormatJPEG}}

	// Act
	err := epub.NewFixedLayoutBuilder().Build(filepath.Join(t.TempDir(), "book.epub"), models.EpubMeta{Title: "x"}, pages)

	// Assert
	assert.Error(t, err)
}
