package documentBuilder

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/pkg"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgaskin/kepubify/v4/kepub"
)

type EpubBuilder struct {
	settings                                                            *domain.Settings
	uuid                                                                uuid.UUID
	ncxBuilder, navBuilder, navElemsBuilder, opfBuilder, opfRefsBuilder *pkg.FileBuilder
	writer                                                              *zip.Writer
	buf                                                                 *bytes.Buffer
	paths                                                               struct {
		text   string
		images string
	}
	volume   *domain.Volume
	pageSide string
}

func (b *EpubBuilder) Copy() BuilderI {
	copy := *b
	return &copy
}

func (b *EpubBuilder) SetSettings(settings *domain.Settings) BuilderI {
	b.settings = settings
	return b
}

func (b *EpubBuilder) Start(volume *domain.Volume) BuilderI {
	pageSide := "right"
	if b.settings.RightToLeft {
		pageSide = "left"
	}

	buf := new(bytes.Buffer)
	*b = EpubBuilder{
		settings: b.settings,
		uuid:     uuid.New(),
		buf:      buf,
		writer:   zip.NewWriter(buf),
		volume:   volume,
		pageSide: pageSide,
		paths: struct{ text, images string }{
			text:   filepath.Join("OEBPS", "Text"),
			images: filepath.Join("OEBPS", "Images"),
		},
	}

	b.startBuilders()
	b.addHeaders()
	b.addFile(
		filepath.Join("META-INF", "container.xml"),
		XML,
	)
	b.addStyles()

	return b
}

func (b *EpubBuilder) Build() (string, error) {
	b.copyFile(
		b.volume.Chapters[0].Pages[0].Parts[0].Path,
		filepath.Join(
			b.paths.images,
			"cover.jpg",
		),
	)

	for _, chapter := range b.volume.Chapters {
		path := chapter.Pages[0].Parts[0].Path
		ext := filepath.Ext(path)
		partName := strings.TrimSuffix(filepath.Base(path), ext)

		folder := filepath.Join(
			"Text",
			chapter.Name,
			partName+".xhtml",
		)

		// NCX
		b.ncxBuilder.AddFromTemplate(
			NCXNavPoint,
			strings.Replace(folder, string(filepath.Separator), "_", -1),
			chapter.Name,
			filepath.ToSlash(folder),
		)

		// NAV
		b.navElemsBuilder.AddFromTemplate(
			NAVLiElem,
			filepath.ToSlash(folder),
			chapter.Name,
		)
	}

	if err := b.closeBuilders(); err != nil {
		return "", err
	}

	if err := b.writer.Close(); err != nil {
		return "", err
	}

	if b.settings.Profile.IsKepub {
		return b.ConvertToKepub()
	}

	path := filepath.Join(b.settings.Output.Base, b.volume.Name+".epub")
	out, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer out.Close()

	if _, err := io.Copy(out, b.buf); err != nil {
		return "", err
	}

	return path, nil
}

func (b *EpubBuilder) AddPage(page *domain.Page) BuilderI {
	for _, part := range page.Parts {
		b.copyFile(
			part.Path,
			filepath.Join(
				b.paths.images,
				part.ChapterName,
				part.Name,
			),
		)
		b.addHTML(part, page.GetCSSBgStyle())

		nonExtPath := filepath.Join(part.ChapterName, part.Name)
		id := strings.Replace(nonExtPath, string(filepath.Separator), "_", -1)

		switch part.PathOrder {
		case 'X', 'D': // Normal & Rotated
			b.pageSide = pkg.Toggle(b.pageSide, "right", "left")
		case 'B': // Double splitted left
			b.pageSide = "left"
			if b.settings.RightToLeft {
				b.pageSide = "right"
			}
		case 'C': // Double splitted right
			b.pageSide = "right"
			if b.settings.RightToLeft {
				b.pageSide = "left"
			}
		}

		b.opfRefsBuilder.AddFromTemplate(
			OPFItemRef,
			id,
			b.pageSide,
		)

		b.opfBuilder.
			AddFromTemplate(
				OPFItem,
				"page_"+id,
				filepath.ToSlash(filepath.Join("Text", filepath.ToSlash(nonExtPath)+".xhtml")),
				"application/xhtml+xml",
			).
			AddFromTemplate(
				OPFItem,
				"img_"+id,
				filepath.ToSlash(filepath.Join("Images", filepath.ToSlash(nonExtPath)+".jpg")),
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

	w, err := b.writer.Create(filepath.ToSlash(dstPath))
	if err != nil {
		return err
	}

	_, err = io.Copy(w, src)
	return err
}

func (b *EpubBuilder) startBuilders() {
	b.ncxBuilder = pkg.NewFileBuilder().AddFromTemplate(
		NCXStart,
		b.uuid,
		b.volume.Name,
	)

	b.navBuilder = pkg.NewFileBuilder().AddFromTemplate(
		NAVStart,
		b.volume.Name,
	)

	b.opfBuilder = pkg.NewFileBuilder().
		AddFromTemplate(
			OPFStart,
			b.volume.Name,
			b.uuid,
			"0.0",
		).
		AddFromTemplate(
			OPFMetas,
			time.Now().UTC().Format(time.RFC3339),
			b.settings.Profile.Width,
			b.settings.Profile.Height,
			b.settings.GetWritingMode(),
			false,
		)

	b.opfRefsBuilder = pkg.NewFileBuilder()
}

func (b *EpubBuilder) closeBuilders() error {
	b.ncxBuilder.AddFromTemplate(NCXEnd)
	if err := b.ncxBuilder.BuildToZip(b.writer, filepath.Join("OEBPS", "toc.NCX")); err != nil {
		return err
	}

	navLiElems := b.navElemsBuilder.Build()
	b.navBuilder.
		Add(navLiElems).
		Add(NAVBetweenList).
		Add(navLiElems).
		Add(NAVEnd)
	if err := b.navBuilder.BuildToZip(b.writer, filepath.Join("OEBPS", "nav.xhtml")); err != nil {
		return err
	}

	b.opfBuilder.
		AddFromTemplate(
			OPFItem,
			"css",
			filepath.ToSlash(filepath.Join("Text", "style.css")),
			"text/css",
		).
		AddFromTemplate(
			OPFPageProgression,
			b.settings.GetPageProgression(),
		).
		Add(b.opfRefsBuilder.Build()).
		Add(OPFEnd)

	if err := b.opfBuilder.BuildToZip(b.writer, filepath.Join("OEBPS", "content.opf")); err != nil {
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
	_, err := b.addFile(
		filepath.Join(b.paths.text, "style.css"),
		Styles,
	)

	return err
}

func (b *EpubBuilder) ConvertToKepub() (string, error) {
	kPath := filepath.Join(b.settings.Output.Base, b.volume.Name+".kepub.epub")
	out, err := os.Create(kPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	in, err := zip.NewReader(bytes.NewReader(b.buf.Bytes()), int64(b.buf.Len()))
	if err != nil {
		return "", err
	}

	converter := kepub.NewConverter()
	ctx := context.Background()

	return kPath, converter.Convert(ctx, out, in)
}

func (b *EpubBuilder) addHTML(part *domain.PagePart, CSSBgStyle string) error {
	relPath := filepath.ToSlash(filepath.Join("..", "..", "Images", part.ChapterName, part.Name))

	builder := pkg.NewFileBuilder().
		AddFromTemplate(
			HTMLStart,
			part.Name,
			part.Width,
			part.Height,
			CSSBgStyle,
			part.GetTopMargin(b.settings.Profile.Height),
		).
		AddFromTemplate(
			HTMLImg,
			part.Width,
			part.Height,
			relPath,
		)

	zipPath := filepath.Join(b.paths.text, part.ChapterName, part.Name+".xhtml")
	return builder.BuildToZip(b.writer, zipPath)
}
