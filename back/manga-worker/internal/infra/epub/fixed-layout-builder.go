package epub

import (
	"archive/zip"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

var _ ports.IMangaEpubBuilder = (*FixedLayoutBuilder)(nil)

var errNoPages = errors.New("epub without pages")

const (
	pageSideLeft  = "left"
	pageSideRight = "right"
)

// FixedLayoutBuilder writes a pre-paginated (comic) epub with one image per page.
type FixedLayoutBuilder struct{}

func NewFixedLayoutBuilder() *FixedLayoutBuilder { return &FixedLayoutBuilder{} }

func (f *FixedLayoutBuilder) Build(outPath string, meta models.EpubMeta, pages []models.EpubPage) (err error) {
	if len(pages) == 0 {
		return errNoPages
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create epub: %w", err)
	}
	defer func() {
		err = errors.Join(err, out.Close())
	}()

	w := &bookWriter{zip: zip.NewWriter(out), meta: meta, title: html.EscapeString(meta.Title)}
	if err := w.write(pages); err != nil {
		return err
	}
	return w.zip.Close()
}

type bookWriter struct {
	zip   *zip.Writer
	meta  models.EpubMeta
	title string
}

func (w *bookWriter) write(pages []models.EpubPage) error {
	mimetype, err := w.zip.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		return err
	}
	if _, err := io.WriteString(mimetype, "application/epub+zip"); err != nil {
		return err
	}
	if err := w.writeString("META-INF/container.xml", containerXML); err != nil {
		return err
	}
	if err := w.writeString("OEBPS/Text/style.css", styles); err != nil {
		return err
	}

	cover := pages[0]
	coverHref := "Images/cover" + cover.Format.Ext()
	if err := w.copyFile(cover.Path, "OEBPS/"+coverHref); err != nil {
		return fmt.Errorf("write cover: %w", err)
	}

	bookId := uuid.NewString()
	var manifest, spine strings.Builder
	side := pageSideRight
	if w.meta.RTL {
		side = pageSideLeft
	}

	for i, page := range pages {
		id := fmt.Sprintf("p%04d", i+1)
		imageHref := "Images/" + id + page.Format.Ext()
		textHref := "Text/" + id + ".xhtml"

		if err := w.copyFile(page.Path, "OEBPS/"+imageHref); err != nil {
			return fmt.Errorf("write page %d: %w", i+1, err)
		}
		xhtml := fmt.Sprintf(pageXHTML,
			id, page.Width, page.Height, topMargin(page.Height, w.meta.Height),
			page.Width, page.Height, "../"+imageHref,
		)
		if err := w.writeString("OEBPS/"+textHref, xhtml); err != nil {
			return err
		}

		side = nextSide(side, page.Split)
		fmt.Fprintf(&manifest, opfItem, "page_"+id, textHref, "application/xhtml+xml")
		fmt.Fprintf(&manifest, opfItem, "img_"+id, imageHref, page.Format.MediaType())
		fmt.Fprintf(&spine, opfItemRef, "page_"+id, side)
	}

	firstPage := "Text/p0001.xhtml"
	if err := w.writeString("OEBPS/toc.ncx", fmt.Sprintf(ncx, bookId, w.title, firstPage)); err != nil {
		return err
	}
	if err := w.writeString("OEBPS/nav.xhtml", fmt.Sprintf(nav, w.title, firstPage)); err != nil {
		return err
	}

	writingMode, progression := "horizontal-lr", "ltr"
	if w.meta.RTL {
		writingMode, progression = "horizontal-rl", "rtl"
	}
	var opf strings.Builder
	fmt.Fprintf(&opf, opfStart,
		w.title, bookId, time.Now().UTC().Format(time.RFC3339),
		w.meta.Width, w.meta.Height, writingMode, coverHref, cover.Format.MediaType(),
	)
	opf.WriteString(manifest.String())
	fmt.Fprintf(&opf, opfSpineStart, progression)
	opf.WriteString(spine.String())
	opf.WriteString(opfEnd)
	return w.writeString("OEBPS/content.opf", opf.String())
}

// nextSide returns the spread side of a page: halves of a split spread keep their own side and
// whole pages alternate.
func nextSide(previous string, split models.SplitOperation) string {
	switch split {
	case models.SplitToLeft:
		return pageSideLeft
	case models.SplitToRight:
		return pageSideRight
	default:
		if previous == pageSideLeft {
			return pageSideRight
		}
		return pageSideLeft
	}
}

func topMargin(pageHeight, screenHeight int) float64 {
	if screenHeight <= 0 || pageHeight >= screenHeight {
		return 0
	}
	margin := float64(screenHeight-pageHeight) / 2 / float64(screenHeight) * 100
	return math.Round(margin*10) / 10
}

func (w *bookWriter) writeString(name, content string) error {
	file, err := w.zip.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(file, content)
	return err
}

func (w *bookWriter) copyFile(srcPath, name string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := w.zip.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}
