package epub

import (
	"fmt"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/domain/manga"
	"path"
	"strings"
)

type mergedItem struct {
	ID         string
	Href       string
	MediaType  string
	Properties string
}

type navEntry struct {
	Title string
	Href  string
}

type EpubMerger struct {
	settings *manga.ImageSettings
}

func NewEpubMerger() *EpubMerger {
	return &EpubMerger{}
}

func (m *EpubMerger) SetSettings(settings *manga.ImageSettings) *EpubMerger {
	m.settings = settings
	return m
}

func (m *EpubMerger) Merge(paths []*convert.TransactionResultFile, title, author, outputPath string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no se han indicado epubs a fusionar")
	}

	sources := make([]*sourceEpub, 0, len(paths))
	for i, p := range paths {
		src, err := loadSourceEpub(p.Path, i)
		if err != nil {
			return err
		}
		sources = append(sources, src)
	}

	outFiles := make(map[string][]byte)
	var manifest []mergedItem
	var spineIDs []string
	var navEntries []navEntry
	var coverItemID string

	for _, src := range sources {
		bookDir := fmt.Sprintf("b%d", src.index)
		coverID := detectCoverItemID(src.opf)
		oldToNew := make(map[string]mergedItem, len(src.opf.Manifest.Items))

		for _, item := range src.opf.Manifest.Items {
			if item.MediaType == "application/x-dtbncx+xml" || strings.Contains(item.Properties, "nav") {
				continue
			}

			origZipPath := path.Clean(path.Join(src.opfDir, item.Href))
			data, ok := src.files[origZipPath]
			if !ok {
				continue
			}

			newHref := path.Clean(path.Join(bookDir, item.Href))
			newZipPath := path.Join("OEBPS", newHref)
			outFiles[newZipPath] = data

			props := removeToken(item.Properties, "cover-image")
			if src.index == 0 && item.ID == coverID {
				props = strings.TrimSpace(props + " cover-image")
			}

			mi := mergedItem{
				ID:         fmt.Sprintf("b%d_%s", src.index, sanitizeID(item.ID)),
				Href:       newHref,
				MediaType:  item.MediaType,
				Properties: strings.TrimSpace(props),
			}
			oldToNew[item.ID] = mi
			manifest = append(manifest, mi)

			if src.index == 0 && item.ID == coverID {
				coverItemID = mi.ID
			}
		}

		for _, ir := range src.opf.Spine.ItemRefs {
			if mi, ok := oldToNew[ir.IDRef]; ok {
				spineIDs = append(spineIDs, mi.ID)
			}
		}

		bookTitle := strings.TrimSpace(src.opf.Metadata.Title)
		if bookTitle == "" {
			bookTitle = src.baseName
		}
		var firstHref string
		if len(src.opf.Spine.ItemRefs) > 0 {
			if mi, ok := oldToNew[src.opf.Spine.ItemRefs[0].IDRef]; ok {
				firstHref = mi.Href
			}
		}
		navEntries = append(navEntries, navEntry{Title: bookTitle, Href: firstHref})
	}

	if coverItemID == "" {
		return fmt.Errorf("no se pudo determinar la portada del primer epub (%s)", paths[0].Path)
	}

	manifest = append(manifest,
		mergedItem{ID: "nav", Href: "nav.xhtml", MediaType: "application/xhtml+xml", Properties: "nav"},
		mergedItem{ID: "ncx", Href: "toc.ncx", MediaType: "application/x-dtbncx+xml"},
	)

	pageProgression := "ltr"
	writingMode := "horizontal-lr"
	if m.settings != nil && m.settings.RightToLeft {
		pageProgression = "rtl"
		writingMode = "horizontal-rl"
	}

	outFiles["mimetype"] = []byte("application/epub+zip")
	outFiles["META-INF/container.xml"] = []byte(MergerContainerXML)
	outFiles["OEBPS/content.opf"] = []byte(buildOPF(title, author, manifest, spineIDs, coverItemID, "ncx", pageProgression, writingMode))
	outFiles["OEBPS/nav.xhtml"] = []byte(buildNav(title, navEntries))
	outFiles["OEBPS/toc.ncx"] = []byte(buildNCX(title, navEntries))

	return writeEpubZip(outputPath, outFiles)
}
