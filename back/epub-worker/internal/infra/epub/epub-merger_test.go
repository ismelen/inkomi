package epub_test

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/epub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func merge(t *testing.T, meta models.EpubMeta, books ...models.EpubBook) mergedEpub {
	t.Helper()

	out := filepath.Join(t.TempDir(), "merged.epub")
	require.NoError(t, epub.NewEpubMerger().Merge(out, meta, books))
	return readMerged(t, out)
}

func TestEpubMerger_Merge_NoBooks_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	err := epub.NewEpubMerger().Merge(filepath.Join(t.TempDir(), "out.epub"), models.EpubMeta{Title: "x"}, nil)

	// Assert
	assert.Error(t, err)
}

func TestEpubMerger_Merge_InvalidSource_ReturnsErrorAndNoOutput(t *testing.T) {
	t.Parallel()

	// Arrange
	out := filepath.Join(t.TempDir(), "out.epub")
	bad := filepath.Join(t.TempDir(), "bad.epub")
	require.NoError(t, os.WriteFile(bad, []byte("not a zip"), 0o600))

	// Act
	err := epub.NewEpubMerger().Merge(out, models.EpubMeta{Title: "x"}, []models.EpubBook{{Path: bad}})

	// Assert
	assert.Error(t, err)
	assert.NoFileExists(t, out)
}

func TestEpubMerger_Merge_TwoBooks_ProducesValidEpub(t *testing.T) {
	t.Parallel()

	// Arrange
	first := writeFakeEpub(t, fakeEpub{title: "Inner 1", pages: 2, cover: true})
	second := writeFakeEpub(t, fakeEpub{title: "Inner 2", pages: 1, cover: true})

	// Act
	m := merge(t, models.EpubMeta{Title: "Manga Ch.[1 - 2]", Author: "Me"},
		models.EpubBook{Path: first, Title: "Chapter 1"},
		models.EpubBook{Path: second, Title: "Chapter 2"},
	)

	// Assert
	assert.Equal(t, "mimetype", m.first.Name)
	assert.Equal(t, zip.Store, m.first.Method)
	assert.Equal(t, "application/epub+zip", m.files["mimetype"])
	for _, name := range []string{"META-INF/container.xml", "OEBPS/nav.xhtml", "OEBPS/toc.ncx"} {
		assert.NoError(t, xml.Unmarshal([]byte(m.files[name]), new(any)), name)
	}

	assert.Equal(t, "Manga Ch.[1 - 2]", m.opf.Metadata.Title)
	assert.Equal(t, "ja", m.opf.Metadata.Language)
	for _, item := range m.opf.Manifest {
		assert.Contains(t, m.files, path.Join("OEBPS", item.Href), item.Id)
	}

	var spine []string
	for _, ref := range m.opf.Spine.ItemRefs {
		spine = append(spine, ref.IdRef)
	}
	assert.Equal(t, []string{"b0_page0", "b0_page1", "b1_page0"}, spine)

	assert.Contains(t, m.files["OEBPS/nav.xhtml"], `<a href="b0/OEBPS/Text/p0.xhtml">Chapter 1</a>`)
	assert.Contains(t, m.files["OEBPS/nav.xhtml"], `<a href="b1/OEBPS/Text/p0.xhtml">Chapter 2</a>`)
	assert.Contains(t, m.files["OEBPS/toc.ncx"], "<text>Chapter 2</text>")

	uuidRe := regexp.MustCompile(`urn:uuid:([0-9a-f-]+)`)
	assert.Equal(t, uuidRe.FindStringSubmatch(m.files["OEBPS/content.opf"])[1], uuidRe.FindStringSubmatch(m.files["OEBPS/toc.ncx"])[1])
}

func TestEpubMerger_Merge_Direction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		rtl         bool
		direction   string
		writingMode string
	}{
		{"RTL", true, "rtl", "horizontal-rl"},
		{"LTR", false, "ltr", "horizontal-lr"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			src := writeFakeEpub(t, fakeEpub{title: "Inner", pages: 1, fixedLayout: true})

			// Act
			m := merge(t, models.EpubMeta{Title: "Manga", RTL: tc.rtl}, models.EpubBook{Path: src})

			// Assert
			assert.Equal(t, tc.direction, m.opf.Spine.Direction)
			mode, ok := m.meta("primary-writing-mode")
			assert.True(t, ok)
			assert.Equal(t, tc.writingMode, mode)
			assert.Equal(t, 1, strings.Count(m.files["OEBPS/content.opf"], "primary-writing-mode"))
		})
	}
}

func TestEpubMerger_Merge_FixedLayout_KeepsRenditionMetadata(t *testing.T) {
	t.Parallel()

	// Arrange
	src := writeFakeEpub(t, fakeEpub{title: "Inner", pages: 2, fixedLayout: true})

	// Act
	m := merge(t, models.EpubMeta{Title: "Manga", RTL: true}, models.EpubBook{Path: src})

	// Assert
	assert.Equal(t, "rendition: http://www.idpf.org/2013/rendition", m.opf.Prefix)
	for key, expected := range map[string]string{
		"rendition:layout": "pre-paginated",
		"rendition:spread": "landscape",
		"book-type":        "comic",
	} {
		got, ok := m.meta(key)
		assert.True(t, ok, key)
		assert.Equal(t, expected, got, key)
	}
	_, refined := m.meta("title-type")
	assert.False(t, refined, "metas refining source ids must be dropped")
	assert.Equal(t, "page-spread-right", m.opf.Spine.ItemRefs[0].Properties)
	assert.Equal(t, "page-spread-left", m.opf.Spine.ItemRefs[1].Properties)
}

func TestEpubMerger_Merge_Cover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		firstCover    bool
		secondCover   bool
		expectedCover string
	}{
		{"From first book", true, true, "b0_cover"},
		{"From second book", false, true, "b1_cover"},
		{"None", false, false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			first := writeFakeEpub(t, fakeEpub{title: "1", pages: 1, cover: tc.firstCover})
			second := writeFakeEpub(t, fakeEpub{title: "2", pages: 1, cover: tc.secondCover})

			// Act
			m := merge(t, models.EpubMeta{Title: "Manga"}, models.EpubBook{Path: first}, models.EpubBook{Path: second})

			// Assert
			cover, ok := m.meta("cover")
			assert.Equal(t, tc.expectedCover != "", ok)
			assert.Equal(t, tc.expectedCover, cover)

			var coverImages []string
			for _, item := range m.opf.Manifest {
				if item.Properties == "cover-image" {
					coverImages = append(coverImages, item.Id)
				}
			}
			if tc.expectedCover == "" {
				assert.Empty(t, coverImages)
			} else {
				assert.Equal(t, []string{tc.expectedCover}, coverImages)
			}
		})
	}
}

func TestEpubMerger_Merge_ImagesOutsideOPFDir_KeepsLinksInsideBookDir(t *testing.T) {
	t.Parallel()

	// Arrange
	first := writeFakeEpub(t, fakeEpub{title: "1", pages: 1, imagesOutsideOPF: true})
	second := writeFakeEpub(t, fakeEpub{title: "2", pages: 1, imagesOutsideOPF: true})

	// Act
	m := merge(t, models.EpubMeta{Title: "Manga"}, models.EpubBook{Path: first}, models.EpubBook{Path: second})

	// Assert
	page := "OEBPS/b1/OEBPS/Text/p0.xhtml"
	require.Contains(t, m.files, page)
	src := regexp.MustCompile(`src="([^"]+)"`).FindStringSubmatch(m.files[page])[1]
	assert.Contains(t, m.files, path.Join(path.Dir(page), src))
	assert.Equal(t, "OEBPS/b1/Images/p0.jpg", path.Join(path.Dir(page), src))
}

func TestEpubMerger_Merge_TocTitleFallbacks(t *testing.T) {
	t.Parallel()

	// Arrange
	first := writeFakeEpub(t, fakeEpub{title: "Inner title", pages: 1})
	second := writeFakeEpub(t, fakeEpub{title: "", pages: 1})

	// Act
	m := merge(t, models.EpubMeta{Title: "Manga"}, models.EpubBook{Path: first}, models.EpubBook{Path: second})

	// Assert
	assert.Contains(t, m.files["OEBPS/nav.xhtml"], ">Inner title</a>")
	assert.Contains(t, m.files["OEBPS/nav.xhtml"], ">Part 2</a>")
}
