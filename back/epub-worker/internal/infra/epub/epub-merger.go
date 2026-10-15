package epub

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
)

var errNoBooks = errors.New("no epubs to merge")

const (
	ncxMediaType = "application/x-dtbncx+xml"
	navProperty  = "nav"
)

var generatedFiles = []struct{ name, template string }{
	{"OEBPS/content.opf", "opf"},
	{"OEBPS/nav.xhtml", "nav"},
	{"OEBPS/toc.ncx", "ncx"},
}

// skippedMetas are regenerated for the merged book instead of inherited from the first source.
var skippedMetas = map[string]bool{"cover": true, "primary-writing-mode": true, "dcterms:modified": true}

type manifestItem struct {
	Id         string
	Href       string
	MediaType  string
	Properties string
}

type tocEntry struct {
	Title string
	Href  string
}

type packageData struct {
	UUID        string
	Title       string
	Author      string
	Language    string
	Modified    string
	Prefix      string
	CoverId     string
	WritingMode string
	Direction   string
	Metas       []opfMeta
	Manifest    []manifestItem
	Spine       []opfItemRef
	Toc         []tocEntry
}

type EpubMerger struct{}

func NewEpubMerger() *EpubMerger {
	return &EpubMerger{}
}

// Merge writes one epub with every book under OEBPS/b<N>/, keeping their original tree so
// relative links still resolve. With meta.RTL the books go in reverse order; the cover and the
// rendition metadata come from the first book of the resulting order.
func (m *EpubMerger) Merge(outputPath string, meta models.EpubMeta, books []models.EpubBook) (err error) {
	if len(books) == 0 {
		return errNoBooks
	}
	if meta.RTL {
		books = slices.Clone(books)
		slices.Reverse(books)
	}

	sources := make([]*sourceEpub, 0, len(books))
	defer func() {
		for _, src := range sources {
			// sources are only read; a close failure can't corrupt the output
			_ = src.Close()
		}
	}()
	for _, book := range books {
		src, err := openSource(book.Path)
		if err != nil {
			return err
		}
		sources = append(sources, src)
	}

	w, err := newEpubWriter(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.Close())
		if err != nil {
			// the partial output is useless; the merge error is what matters
			_ = os.Remove(outputPath)
		}
	}()

	pkg := newPackageData(meta, sources[0])
	for i, src := range sources {
		if err := pkg.addBook(w, i, books[i], src); err != nil {
			return err
		}
	}

	if err := w.write("META-INF/container.xml", containerXML); err != nil {
		return err
	}
	for _, file := range generatedFiles {
		if err := w.render(file.name, file.template, pkg); err != nil {
			return err
		}
	}
	return nil
}

func newPackageData(meta models.EpubMeta, first *sourceEpub) *packageData {
	pkg := &packageData{
		UUID:        uuid.NewString(),
		Title:       meta.Title,
		Author:      meta.Author,
		Language:    cmp.Or(strings.TrimSpace(first.opf.Metadata.Language), "und"),
		Modified:    time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Prefix:      first.opf.Prefix,
		WritingMode: "horizontal-lr",
		Direction:   "ltr",
	}
	if meta.RTL {
		pkg.WritingMode, pkg.Direction = "horizontal-rl", "rtl"
	}

	for _, m := range first.opf.Metadata.Metas {
		if m.Refines != "" || skippedMetas[m.Name] || skippedMetas[m.Property] || (m.Name == "" && m.Property == "") {
			continue
		}
		m.Value = strings.TrimSpace(m.Value)
		pkg.Metas = append(pkg.Metas, m)
	}
	return pkg
}

func (p *packageData) addBook(w *epubWriter, index int, book models.EpubBook, src *sourceEpub) error {
	dir := fmt.Sprintf("b%d", index)
	coverId := ""
	if p.CoverId == "" {
		coverId = src.coverItemId()
	}

	ids := make(map[string]manifestItem, len(src.opf.Manifest.Items))
	for _, item := range src.opf.Manifest.Items {
		if item.MediaType == ncxMediaType || hasToken(item.Properties, navProperty) {
			continue
		}
		f, ok := src.entry(item.Href)
		if !ok {
			continue
		}
		if err := w.copy(f, path.Join("OEBPS", dir, f.Name)); err != nil {
			return err
		}

		merged := manifestItem{
			Id:         fmt.Sprintf("%s_%s", dir, sanitizeId(item.Id)),
			Href:       path.Join(dir, src.opfDir, item.Href),
			MediaType:  item.MediaType,
			Properties: removeToken(item.Properties, "cover-image"),
		}
		if item.Id == coverId && strings.HasPrefix(item.MediaType, "image/") {
			merged.Properties = addToken(merged.Properties, "cover-image")
			p.CoverId = merged.Id
		}
		ids[item.Id] = merged
		p.Manifest = append(p.Manifest, merged)
	}

	firstHref := ""
	for _, ref := range src.opf.Spine.ItemRefs {
		merged, ok := ids[ref.IdRef]
		if !ok {
			continue
		}
		if firstHref == "" {
			firstHref = merged.Href
		}
		ref.IdRef = merged.Id
		p.Spine = append(p.Spine, ref)
	}

	if firstHref != "" {
		title := cmp.Or(strings.TrimSpace(book.Title), strings.TrimSpace(src.opf.Metadata.Title), fmt.Sprintf("Part %d", index+1))
		p.Toc = append(p.Toc, tocEntry{Title: title, Href: firstHref})
	}
	return nil
}
