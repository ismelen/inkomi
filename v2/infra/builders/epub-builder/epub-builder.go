package epubBuilder

import (
	"archive/zip"
	"fmt"
	"io"
	"ismelen/ermc/internal/pkg"
	"ismelen/ermc/v2/domain"
	fileBuilder "ismelen/ermc/v2/infra/builders/file-builder"
	"ismelen/ermc/v2/infra/image"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EpubBuilder struct {
	settings *domain.ImageSettings
	profile *domain.Profile
	builders builders
	writer *zip.Writer
	out *os.File
	name string
	pageSide string
	hasCover bool
	mu sync.Mutex
}

func New() *EpubBuilder { return &EpubBuilder{} }

type builders struct {
	ncx, nav, navElems, opf, opfRefs *fileBuilder.FileBuilder
}

var PATHS = struct{ text, images string }{
	text:   filepath.Join("OEBPS", "Text"),
	images: filepath.Join("OEBPS", "Images"),
}

const (
	PAGE_RIGHT = "right"
	PAGE_LEFT = "left"
)

func (b *EpubBuilder) SetSettings(settings *domain.ImageSettings, profile *domain.Profile) *EpubBuilder {
	b.settings = settings
	b.profile = profile
	return b
}

func (b *EpubBuilder) Start(name string, outDir string) *EpubBuilder {
	pageSide := PAGE_RIGHT
	if b.settings.RightToLeft {
		pageSide = PAGE_LEFT
	}

	path := filepath.Join(outDir, name+".epub")
	out, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return b
	}

	*b = EpubBuilder{
		settings: b.settings,
		out:      out,
		writer:   zip.NewWriter(out),
		name:   name,
		pageSide: pageSide,
		mu: sync.Mutex{},
	}

	b.startBuilders()
	b.addHeaders()
	b.addFile(
		filepath.ToSlash(filepath.Join("META-INF", "container.xml")),
		XML,
	)
	b.addStyles()

	return b
}

func (b *EpubBuilder) Build() (string, error) {
	if err := b.closeBuilders(); err != nil {
		return "", err
	}

	if err := b.writer.Close(); err != nil {
		return "", err
	}

	b.out.Close()
	
	// if b.profile.IsKepub {
	// 	return b.ConvertToKepub()
	// }
	
	return b.out.Name(), nil
}

func (b *EpubBuilder) AddPage(page *domain.Page, fstPage bool) *EpubBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.hasCover {
		part := page.Parts[0]
		b.copyFile(
			part.Path,
			filepath.ToSlash(filepath.Join(
				PATHS.images,
				"cover.jpg",
			)),
		)
		b.hasCover = true
	}

	if fstPage {
		part := page.Parts[0]

		folder := filepath.Join(
			"Text",
			part.ChapterName,
			part.Name+".xhtml",
		)

		b.builders.ncx.AddFromTemplate(
			NCXNavPoint,
			strings.Replace(folder, string(filepath.Separator), "_", -1),
			part.ChapterName,
			filepath.ToSlash(folder),
		)

		b.builders.navElems.AddFromTemplate(
			NAVLiElem,
			filepath.ToSlash(folder),
			part.ChapterName,
		)
	}
	
	for _, part := range page.Parts {
		path := filepath.Join(
			PATHS.images,
			part.ChapterName,
			part.Name,
		)
		b.copyFile(
			part.Path,
			filepath.ToSlash(path+part.Ext),
		)
		b.addHTML(part, page.GetCSSBgStyle())

		nonExtPath := filepath.Join(part.ChapterName, part.Name)
		id := strings.Replace(nonExtPath, string(filepath.Separator), "_", -1)

		switch part.Split {
		case image.None, image.Rotated:
			b.pageSide = pkg.Toggle(b.pageSide, PAGE_RIGHT, PAGE_LEFT)
		case image.ToLeft: 
			b.pageSide = PAGE_LEFT
		case image.ToRight:
			b.pageSide = PAGE_RIGHT 
		}

		b.builders.opfRefs.AddFromTemplate(
			OPFItemRef,
			id,
			b.pageSide,
		)

		b.builders.opf.
			AddFromTemplate(
				OPFItem,
				"page_"+id,
				filepath.ToSlash(filepath.Join("Text", nonExtPath+".xhtml")),
				"application/xhtml+xml",
			).
			AddFromTemplate(
				OPFItem,
				"img_"+id,
				filepath.ToSlash(filepath.Join("Images", nonExtPath+".jpg")),
				"image/jpeg",
			)
	}

	return b
}

func (b *EpubBuilder) copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	w, err := b.writer.Create(dstPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, src)
	return err
}

func (b *EpubBuilder) startBuilders() {
	uuid := uuid.New()

	b.builders.ncx = fileBuilder.New().AddFromTemplate(
		NCXStart,
		uuid,
		b.name,
	)

	b.builders.nav = fileBuilder.New().AddFromTemplate(
		NAVStart,
		b.name,
	)

	writingMode := "horizontal-lr"
	if b.settings.RightToLeft {
		writingMode = "horizontal-rl"
	}

	b.builders.opf = fileBuilder.New().
		AddFromTemplate(
			OPFStart,
			b.name,
			uuid,
			"0.0",
		).
		AddFromTemplate(
			OPFMetas,
			time.Now().UTC().Format(time.RFC3339),
			b.profile.Width,
			b.profile.Height,
			writingMode,
			false,
		)

	b.builders.opfRefs = fileBuilder.New()
	b.builders.navElems = fileBuilder.New()
}

func (b *EpubBuilder) closeBuilders() error {
	b.builders.ncx.AddFromTemplate(NCXEnd)
	if err := b.builders.ncx.BuildToZip(b.writer, filepath.ToSlash(filepath.Join("OEBPS", "toc.ncx"))); err != nil {
		return err
	}

	navLiElems := b.builders.navElems.Build()
	b.builders.nav.
		Add(navLiElems).
		Add(NAVBetweenList).
		Add(navLiElems).
		Add(NAVEnd)
	if err := b.builders.nav.BuildToZip(b.writer, filepath.ToSlash(filepath.Join("OEBPS", "nav.xhtml"))); err != nil {
		return err
	}

	pageProgression := "ltr"
	if b.settings.RightToLeft {
		pageProgression = "rtl"
	}

	b.builders.opf.
		AddFromTemplate(
			OPFItem,
			"css",
			filepath.ToSlash(filepath.Join("Text", "style.css")),
			"text/css",
		).
		AddFromTemplate(
			OPFPageProgression,
			pageProgression,
		).
		Add(b.builders.opfRefs.Build()).
		Add(OPFEnd)

	if err := b.builders.opf.BuildToZip(b.writer, filepath.ToSlash(filepath.Join("OEBPS", "content.opf"))); err != nil {
		return err
	}

	return nil
}

func (b *EpubBuilder) addHeaders() (err error) {
	w, err := b.writer.CreateHeader(MimeHeader)
	if err != nil {
		return
	}

	_, err = w.Write([]byte("application/epub+zip"))
	if err != nil {
		return
	}

	return
}

func (b *EpubBuilder) addFile(zipPath string, content string) (*io.Writer, error) {
	w, err := b.writer.Create(filepath.ToSlash(zipPath))
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (b *EpubBuilder) addStyles() error {
	return fileBuilder.New().
		Add(Styles).
		BuildToZip(
			b.writer, 
			filepath.ToSlash(filepath.Join(PATHS.text, "style.css")),
		)
}


// TODO: KEPUBIFY
// func (b *EpubBuilder) ConvertToKepub() (string, error) {
// 	kPath := filepath.Join(b.settings.Output.Base, b.name+".kepub.epub")
// 	out, err := os.Create(kPath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer out.Close()

// 	in, err := zip.OpenReader(b.out.Name())
// 	if err != nil {
// 		return "", err
// 	}

// 	converter := kepub.NewConverter()
// 	ctx := context.Background()

// 	return kPath, converter.Convert(ctx, out, in)
// }

func (b *EpubBuilder) addHTML(part *domain.PagePart, CSSBgStyle string) error {
	relPath := filepath.ToSlash(filepath.Join("..", "..", "Images", part.ChapterName, part.Name+part.Ext))

	builder := fileBuilder.New().
		AddFromTemplate(
			HTMLStart,
			part.Name,
			part.Width,
			part.Height,
			CSSBgStyle,
			part.GetTopMargin(b.profile.Height),
		).
		AddFromTemplate(
			HTMLImg,
			part.Width,
			part.Height,
			relPath,
		)

	zipPath := filepath.Join(PATHS.text, part.ChapterName, part.Name+".xhtml")
	return builder.BuildToZip(b.writer, filepath.ToSlash(zipPath))
}
