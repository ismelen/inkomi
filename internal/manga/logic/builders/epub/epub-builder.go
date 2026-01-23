package EpubBuilder

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	MangaModels "ismelen/ermc/internal/manga/domain/models"
	ContentBuilder "ismelen/ermc/internal/manga/logic/builders/content"
	EpubTemplates "ismelen/ermc/internal/manga/logic/templates/epub"
	StringUtils "ismelen/ermc/internal/utils/strings"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgaskin/kepubify/v4/kepub"
)

type EpubBuilder struct {
	opts        *MangaModels.ConverterOptions
	chapters    []*MangaModels.Chapter
	dstFileName string
	ncxBuilder *ContentBuilder.ContentBuilder
	navBuilder *ContentBuilder.ContentBuilder
	navElemsBuilder * ContentBuilder.ContentBuilder
	uuid uuid.UUID
}

var imagesPath = filepath.Join("OEBPS", "Images")
var textPath = filepath.Join("OEBPS", "Text")

func New(opts *MangaModels.ConverterOptions, dstFileName string, chapters ...*MangaModels.Chapter) *EpubBuilder {
	return &EpubBuilder{
		opts: opts, 
		chapters: chapters, 
		dstFileName: dstFileName,
		ncxBuilder: ContentBuilder.New(),
		navBuilder: ContentBuilder.New(),
		navElemsBuilder: ContentBuilder.New(),
		uuid: uuid.New(),
	}
}

func (this *EpubBuilder) Build() (string, error) {
	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)

	// this.AddHeaders(z)
	// this.addFile(
	// 	z,
	// 	filepath.Join("META-INF", "container.xml"),
	// 	EpubTemplates.XML,
	// )
	// this.CopyFiles(z)
	// this.AddStyles(z)

	this.StartBuilders()

	for _, chapter := range this.chapters {
		for _, page := range chapter.Pages {
			for i := range page.Count {
				part := page.Parts[i]
				copyFile(
					z,
					part.Path,
					filepath.Join(
						imagesPath,
						chapter.NormalizedName,
						part.Title+".jpg",
					),
				)

				// HTML
				if err := this.addHTML(z, part, page.GetCSSBgStyle()); err != nil {
					return "", err
				}
			}
		}

		// NCX
		folder := filepath.Join(
			"Text",
			chapter.NormalizedName,
			chapter.Pages[0].Parts[0].Title+".xhtml",
		)
		this.ncxBuilder.AddFromTemplate(
			EpubTemplates.NCXNavPoint,
			strings.Replace(folder, string(filepath.Separator), "_", -1),
			chapter.NormalizedName,
			filepath.ToSlash(folder),
		)	

		// NAV
		this.navElemsBuilder.AddFromTemplate(
			EpubTemplates.NAVLiElem,
			filepath.ToSlash(folder),
			chapter.NormalizedName,
		)
	}
	copyFile(
		z,
		this.chapters[0].Pages[0].Parts[0].Path,
		filepath.Join(
			imagesPath,
			"cover.jpg",
		),
	)

	if err := this.CloseBuilders(z); err != nil {
		return "", err
	}
	
	if err := this.addOPF(z); err != nil {
		return "", err
	}

	if err := z.Close(); err != nil {
		return "", err
	}

	if this.opts.ProfileData.IsKepub {
		return this.ConvertToKepub(buf)
	}

	path := filepath.Join(this.opts.Output, this.dstFileName+".epub")
	out, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer out.Close()

	if _, err := io.Copy(out, buf); err != nil {
		return "", err
	}

	return path, nil
}

func (this *EpubBuilder) StartBuilders() {
	this.ncxBuilder.AddFromTemplate(
		EpubTemplates.NCXStart,
		this.uuid,
		this.dstFileName,
	)

	this.navBuilder.AddFromTemplate(
		EpubTemplates.NAVStart,
		this.dstFileName,
	)
}

func (this *EpubBuilder) CloseBuilders(z *zip.Writer) error {
	this.ncxBuilder.AddFromTemplate(EpubTemplates.NCXEnd)
	if err := this.ncxBuilder.BuildToZip(z, filepath.Join("OEBPS", "toc.NCX")); err != nil {
		return err
	}


	navLiElems := this.navElemsBuilder.Build()
	this.navBuilder.
		Add(navLiElems).
		Add(EpubTemplates.NAVBetweenList).
		Add(navLiElems).
		Add(EpubTemplates.NAVEnd)
	if err := this.navBuilder.BuildToZip(z, filepath.Join("OEBPS", "nav.xhtml")); err != nil {
		return err
	}

	return nil
}


func (this *EpubBuilder) ConvertToKepub(buf *bytes.Buffer) (string, error) {
	kPath := filepath.Join(this.opts.Output, this.dstFileName+".kepub.epub")
	out, err := os.Create(kPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	in, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return "", err
	}

	converter := kepub.NewConverter()
	ctx := context.Background()

	return kPath, converter.Convert(ctx, out, in)
}

func (this *EpubBuilder) CopyFiles(z *zip.Writer) error {
	copyFile(
		z,
		this.chapters[0].Pages[0].Parts[0].Path,
		filepath.Join(
			imagesPath,
			"cover.jpg",
		),
	)
	for _, chapter := range this.chapters {
		for _, page := range chapter.Pages {
			for i := range page.Count {
				p := page.Parts[i]
				copyFile(
					z,
					p.Path,
					filepath.Join(
						imagesPath,
						chapter.NormalizedName,
						p.Title+".jpg",
					),
				)
			}
		}
	}

	return nil
}

func copyFile(z *zip.Writer, srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	w, err := z.Create(filepath.ToSlash(dstPath))
	if err != nil {
		return err
	}

	_, err = io.Copy(w, src)
	return err
}

func (this *EpubBuilder) AddHeaders(z *zip.Writer) (err error) {
	w, err := z.CreateHeader(EpubTemplates.MimeHeader)
	if err != nil {
		return
	}

	_, err = w.Write([]byte("application/epub+zip"))
	if err != nil {
		return
	}

	return
}

func (this *EpubBuilder) AddStyles(z *zip.Writer) error {
	_, err := this.addFile(
		z,
		filepath.Join(textPath, "style.css"),
		EpubTemplates.Styles,
	)

	return err
}

func (this *EpubBuilder) addFile(z *zip.Writer, zipPath string, content string) (*io.Writer, error) {
	w, err := z.Create(filepath.ToSlash(zipPath))
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (this *EpubBuilder) addHTML(z *zip.Writer, part *MangaModels.PagePart, CSSBgStyle string) error {
	partName := filepath.Base(part.Path)
	chapterName := filepath.Dir(part.Path)
	relPath := filepath.ToSlash(filepath.Join("..", "..", "Images", chapterName, partName))

	builder := ContentBuilder.New().
		AddFromTemplate(
			EpubTemplates.HTMLStart,
			part.Title,
			part.W,
			part.H,
			CSSBgStyle,
			part.GetTopMargin(this.opts.ProfileData.Height),
		).
		AddFromTemplate(
			EpubTemplates.HTMLImg,
			part.W,
			part.H,
			relPath,
		)

	zipPath := filepath.Join(textPath, chapterName, part.Title+".xhtml")
	return builder.BuildToZip(z, zipPath)
}

func (this *EpubBuilder) addOPF(z *zip.Writer) error {
	elemsBuilder := ContentBuilder.New()
	builder := ContentBuilder.New().
		AddFromTemplate(
			EpubTemplates.OPFStart,
			this.dstFileName,
			this.uuid,
			"0.0",
		).
		AddFromTemplate(
			EpubTemplates.OPFMetas,
			time.Now().UTC().Format(time.RFC3339),
			this.opts.ProfileData.Width,
			this.opts.ProfileData.Height,
			this.opts.GetWritingMode(),
			false,
		)
	
	pageSide := this.opts.GetSpreadShiftPageSide()

	for _, chapter := range this.chapters {
		for _, page := range chapter.Pages {
			for i := range page.Count {
				part := page.Parts[i]
				folder := filepath.Join(
					chapter.NormalizedName,
					part.Title,
				)
				id := strings.Replace(
					folder,
					string(filepath.Separator),
					"_",
					-1,
				)
				var spreadPropery string
				switch part.Mode {
				case 'R', '1', '2':
					spreadPropery = MangaModels.SpreadProperties[part.Mode]
					pageSide = this.opts.GetPageSide()
				default:
					spreadPropery = pageSide
					pageSide = StringUtils.Toggle(pageSide, "right", "left")
				}

				elemsBuilder.AddFromTemplate(
					EpubTemplates.OPFItemRef,
					id,
					spreadPropery,
				)

				builder.AddFromTemplate(
					EpubTemplates.OPFItem,
					"page_"+id,
					filepath.ToSlash(filepath.Join("Text", filepath.ToSlash(folder)+".xhtml")),
					"application/xhtml+xml",
				).
				AddFromTemplate(
					EpubTemplates.OPFItem,
					"img_"+id,
					filepath.ToSlash(filepath.Join("Images", filepath.ToSlash(folder)+".jpg")),
					"image/jpeg",
				)
			}
		}
	}
	
	builder.AddFromTemplate(
		EpubTemplates.OPFItem,
		"css",
		filepath.ToSlash(filepath.Join("Text", "style.css")),
		"text/css",
	).
	AddFromTemplate(
		EpubTemplates.OPFPageProgression,
		this.opts.GetPageProgression(),
	).
	Add(elemsBuilder.Build()).
	Add(EpubTemplates.OPFEnd)

	return builder.BuildToZip(z, filepath.Join("OEBPS", "content.opf"))
}



// func (this *EpubBuilder) BuildOPF(z *zip.Writer, uuid uuid.UUID) error {
// 	path := filepath.Join("OEBPS", "content.opf")

// 	var content string
// 	var refList []string
// 	for _, chapter := range this.chapters {
// 		for _, page := range chapter.Pages {
// 			for i := range page.Count {
// 				folder := filepath.Join(
// 					chapter.NormalizedName,
// 					page.Parts[i].Title,
// 				)
// 				id := strings.Replace(
// 					folder,
// 					string(filepath.Separator),
// 					"_",
// 					-1,
// 				)
// 				refList = append(refList, id)

// 				content += fmt.Sprintf(
// 					EpubTemplates.OPFItem,
// 					"page_"+id,
// 					filepath.ToSlash(filepath.Join("Text", filepath.ToSlash(folder)+".xhtml")),
// 					"application/xhtml+xml",
// 				)

// 				content += fmt.Sprintf(
// 					EpubTemplates.OPFItem,
// 					"img_"+id,
// 					filepath.ToSlash(filepath.Join("Images", filepath.ToSlash(folder)+".jpg")),
// 					"image/jpeg",
// 				)
// 			}
// 		}
// 	}

// 	content += fmt.Sprintf(
// 		EpubTemplates.OPFItem,
// 		"css",
// 		filepath.ToSlash(filepath.Join("Text", "style.css")),
// 		"text/css",
// 	)

// 	var pageSide string
// 	if this.opts.Manga {
// 		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "rtl")
// 		pageSide = "right"
// 	} else {
// 		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "ltr")
// 		pageSide = "left"
// 	}

// 	if this.opts.SpreadShift {
// 		if pageSide == "right" {
// 			pageSide = "left"
// 		} else {
// 			pageSide = "right"
// 		}
// 	}

// 	var pageSpreadPropertyList []string
// 	for _, ref := range refList {
// 		ending := ref[len(ref)-6:]
// 		switch ending {
// 		case "-ermc-a", "-ermc-d":
// 			pageSpreadPropertyList = append(pageSpreadPropertyList, "center")
// 			pageSide = calculatePageSide(this.opts.Manga)
// 		case "-ermc-b":
// 			pageSpreadPropertyList = append(pageSpreadPropertyList, "right")
// 			pageSide = calculatePageSide(this.opts.Manga)
// 		case "-ermc-c":
// 			pageSpreadPropertyList = append(pageSpreadPropertyList, "left")
// 			pageSide = calculatePageSide(this.opts.Manga)
// 		default:
// 			pageSpreadPropertyList = append(pageSpreadPropertyList, pageSide)
// 			if pageSide == "right" {
// 				pageSide = "left"
// 			} else {
// 				pageSide = "right"
// 			}
// 		}
// 	}

// 	spreadSeen := false
// 	for i := len(refList) - 1; i >= 0; i-- {
// 		ref := refList[i]
// 		ending := ref[len(ref)-6:]

// 		if "-ermc-x" != ending {
// 			spreadSeen = true
// 			pageSide = this.opts.GetPageSide()
// 		} else if spreadSeen {
// 			pageSpreadPropertyList[i] = pageSide
// 			StringUtils.Toggle(pageSide, "right", "left")
// 		}
// 	}

// 	for i := 0; i < len(refList); i++ {
// 		content += fmt.Sprintf(
// 			EpubTemplates.OPFItemRef,
// 			refList[i],
// 			pageSpreadPropertyList[i],
// 		)
// 	}

// 	content += EpubTemplates.OPFEnd

// 	_, err := this.addFile(z, path, content)
// 	return err
// }

// func calculatePageSide(manga bool) string {
// 	if manga {
// 		return "right"
// 	}
// 	return "left"
// }
