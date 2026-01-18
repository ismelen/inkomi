package EpubBuilder

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	manga "ismelen/ermc/internal/manga/logic/models"
	EpubTemplates "ismelen/ermc/internal/manga/logic/templates/epub"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgaskin/kepubify/v4/kepub"
)

type EpubBuilder struct {
	opts        *manga.ConverterOptions
	chapters    []*manga.Chapter
	dstFileName string
}

var imagesPath = filepath.Join("OEBPS", "Images")
var textPath = filepath.Join("OEBPS", "Text")

func New(opts *manga.ConverterOptions, dstFileName string, chapters ...*manga.Chapter) *EpubBuilder {
	return &EpubBuilder{opts, chapters, dstFileName}
}

func (this *EpubBuilder) Build() (string, error) {
	// path := filepath.Join(this.opts.Output, this.dstFileName+".epub")
	// out, err := os.Create(path)
	// if err != nil {
	// 	return path, err
	// }
	// defer out.Close()
	// 
	// z := zip.NewWriter(out)
	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)

	this.AddHeaders(z)
	this.addFile(
		z,
		filepath.Join("META-INF", "container.xml"),
		EpubTemplates.XML,
	)
	this.CopyFiles(z)
	this.AddStyles(z)

	for _, chapter := range this.chapters {
		for _, page := range chapter.Pages {
			for i := range page.Count {
				this.BuildHTML(z, page.Parts[i], page, chapter.NormalizedName)
			}
		}
	}

	uuid := uuid.New()

	if err := this.BuildNCX(z, uuid); err != nil {
		return "", err
	}

	if err := this.BuildOPF(z, uuid); err != nil {
		return "", err
	}

	if err := this.BuildNAV(z); err != nil {
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

func (this *EpubBuilder) BuildHTML(z *zip.Writer, payload *manga.PagePart, page *manga.Page, chapterName string) (err error) {
	aditionalStyle := ""
	if !page.HasWhiteBg {
		aditionalStyle = "background-color:#000000;"
	}

	pathParts := strings.Split(payload.Path, string(filepath.Separator))
	relPath := filepath.Join("..", "..", "Images")
	for _, part := range pathParts[len(pathParts)-2:] {
		relPath += string(filepath.Separator) + part
	}

	htmlPath := filepath.Join(textPath, chapterName, payload.Title+".xhtml")

	deviceHeight := this.opts.ProfileData.Height
	imgWidth, imgHeight := payload.W, payload.H

	content := fmt.Sprintf(
		EpubTemplates.HTMLStart,
		payload.Title,
		imgWidth,
		imgHeight,
		aditionalStyle,
		getTopMargin(deviceHeight, imgHeight),
	)

	content += fmt.Sprintf(
		EpubTemplates.HTMLImg,
		imgWidth,
		imgHeight,
		filepath.ToSlash(relPath),
	)

	_, err = this.addFile(z, htmlPath, content)
	return
}

func getTopMargin(deviceHeight, imgHeight int) float64 {
	y := ((deviceHeight - imgHeight) / 2) / deviceHeight * 100
	return math.Round(float64(y*10)) / 10
}

func (this *EpubBuilder) BuildNCX(z *zip.Writer, uuid uuid.UUID) error {
	path := filepath.Join("OEBPS", "toc.ncx")

	content := fmt.Sprintf(
		EpubTemplates.NCXStart,
		uuid,
		this.dstFileName,
	)

	for _, chapter := range this.chapters {
		folder := filepath.Join(
			"Text",
			chapter.NormalizedName,
			chapter.Pages[0].Parts[0].Title+".xhtml",
		)

		content += fmt.Sprintf(
			EpubTemplates.NCXNavPoint,
			strings.Replace(folder, string(filepath.Separator), "_", -1),
			chapter.NormalizedName, // chapter.Title
			filepath.ToSlash(folder),
		)
	}

	content += EpubTemplates.NCXEnd

	_, err := this.addFile(z, path, content)
	return err
}

func (this *EpubBuilder) BuildNAV(z *zip.Writer) error {
	path := filepath.Join("OEBPS", "nav.xhtml")

	content := fmt.Sprintf(
		EpubTemplates.NAVStart,
		this.dstFileName,
	)

	var listContent string

	for _, chapter := range this.chapters {
		folder := filepath.Join(
			"Text",
			chapter.NormalizedName,
			chapter.Pages[0].Parts[0].Title+".xhtml",
		)

		listContent += fmt.Sprintf(
			EpubTemplates.NAVLiElem,
			filepath.ToSlash(folder),
			chapter.NormalizedName, // chapter.Title
		)
	}

	content += listContent +
		EpubTemplates.NAVBetweenList +
		listContent +
		EpubTemplates.NAVEnd

	_, err := this.addFile(z, path, content)
	return err
}

func (this *EpubBuilder) BuildOPF(z *zip.Writer, uuid uuid.UUID) error {
	path := filepath.Join("OEBPS", "content.opf")

	content := fmt.Sprintf(
		EpubTemplates.OPFStart,
		this.dstFileName,
		uuid,
		"0.0",
	)

	writingMode := "horizontal-lr"
	if this.opts.Manga {
		writingMode = "horizontal-rl"
	}

	content += fmt.Sprintf(
		EpubTemplates.OPFMetas,
		time.Now().UTC().Format(time.RFC3339),
		this.opts.ProfileData.Width,
		this.opts.ProfileData.Height,
		writingMode,
		false,
	)

	var refList []string
	for _, chapter := range this.chapters {
		for _, page := range chapter.Pages {
			for i := range page.Count {
				folder := filepath.Join(
					chapter.NormalizedName,
					page.Parts[i].Title,
				)
				id := strings.Replace(
					folder,
					string(filepath.Separator),
					"_",
					-1,
				)
				refList = append(refList, id)

				content += fmt.Sprintf(
					EpubTemplates.OPFItem,
					"page_"+id,
					filepath.ToSlash(filepath.Join("Text", filepath.ToSlash(folder)+".xhtml")),
					"application/xhtml+xml",
				)

				content += fmt.Sprintf(
					EpubTemplates.OPFItem,
					"img_"+id,
					filepath.ToSlash(filepath.Join("Images", filepath.ToSlash(folder)+".jpg")),
					"image/jpeg",
				)
			}
		}
	}

	content += fmt.Sprintf(
		EpubTemplates.OPFItem,
		"css",
		filepath.ToSlash(filepath.Join("Text", "style.css")),
		"text/css",
	)

	var pageSide string
	if this.opts.Manga {
		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "rtl")
		pageSide = "right"
	} else {
		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "ltr")
		pageSide = "left"
	}

	if this.opts.SpreadShift {
		if pageSide == "right" {
			pageSide = "left"
		} else {
			pageSide = "right"
		}
	}

	var pageSpreadPropertyList []string
	for _, ref := range refList {
		ending := ref[len(ref)-6:]
		switch ending {
		case "-ermc-a", "-ermc-d":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "center")
			pageSide = calculatePageSide(this.opts.Manga)
		case "-ermc-b":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "right")
			pageSide = calculatePageSide(this.opts.Manga)
		case "-ermc-c":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "left")
			pageSide = calculatePageSide(this.opts.Manga)
		default:
			pageSpreadPropertyList = append(pageSpreadPropertyList, pageSide)
			if pageSide == "right" {
				pageSide = "left"
			} else {
				pageSide = "right"
			}
		}
	}

	spreadSeen := false
	for i := len(refList) - 1; i >= 0; i-- {
		ref := refList[i]
		ending := ref[len(ref)-6:]

		if "-ermc-x" != ending {
			spreadSeen = true
			if this.opts.Manga {
				pageSide = "left"
			} else {
				pageSide = "right"
			}
		} else if spreadSeen {
			pageSpreadPropertyList[i] = pageSide
			if pageSide == "right" {
				pageSide = "left"
			} else {
				pageSide = "right"
			}
		}
	}

	for i := 0; i < len(refList); i++ {
		content += fmt.Sprintf(
			EpubTemplates.OPFItemRef,
			refList[i],
			pageSpreadPropertyList[i],
		)
	}

	content += EpubTemplates.OPFEnd

	_, err := this.addFile(z, path, content)
	return err
}

func calculatePageSide(manga bool) string {
	if manga {
		return "right"
	}
	return "left"
}
