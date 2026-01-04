package EpubBuilder

import (
	"archive/zip"
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
)

type EpubBuilder struct {
	opts        *manga.Options
	chapters    []*manga.ChapterData
	dstFileName string
}

var imagesPath = filepath.Join("OEBPS", "Images")
var textPath = filepath.Join("OEBPS", "Text")

func New(opts *manga.Options, dstFileName string, chapters ...*manga.ChapterData) *EpubBuilder {
	return &EpubBuilder{opts, chapters, dstFileName}
}

func (t *EpubBuilder) Build() (string, error) {
	path := filepath.Join(t.opts.Output, t.dstFileName+".epub")
	out, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer out.Close()

	z := zip.NewWriter(out)
	defer z.Close()

	t.AddHeaders(z)
	t.addFile(
		z,
		filepath.Join("META-INF", "container.xml"),
		EpubTemplates.XML,
	)
	t.CopyFiles(z)
	t.AddStyles(z)

	for _, chapter := range t.chapters {
		for _, page := range chapter.Pages {
			for _, payload := range page.Payloads {
				t.BuildHTML(z, payload, page, chapter.NormalizedName)
			}
		}
	}

	uuid := uuid.New()

	if err := t.BuildNCX(z, uuid); err != nil {
		return path, err
	}

	if err := t.BuildOPF(z, uuid); err != nil {
		return path, err
	}

	if err := t.BuildNAV(z); err != nil {
		return path, err
	}

	return path, nil
}

func (t *EpubBuilder) CopyFiles(z *zip.Writer) error {
	copyFile(
		z,
		t.chapters[0].Pages[0].Payloads[0].Path,
		filepath.Join(
			imagesPath,
			"cover.jpg",
		),
	)
	for _, chapter := range t.chapters {
		for _, page := range chapter.Pages {
			for _, payload := range page.Payloads {
				copyFile(
					z,
					payload.Path,
					filepath.Join(
						imagesPath,
						chapter.NormalizedName,
						payload.Title+".jpg",
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

func (t *EpubBuilder) AddHeaders(z *zip.Writer) (err error) {
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

func (t *EpubBuilder) AddStyles(z *zip.Writer) error {
	_, err := t.addFile(
		z,
		filepath.Join(textPath, "style.css"),
		EpubTemplates.Styles,
	)

	return err
}

func (t *EpubBuilder) addFile(z *zip.Writer, zipPath string, content string) (*io.Writer, error) {
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

func (t *EpubBuilder) BuildHTML(z *zip.Writer, payload *manga.PagePayload, page *manga.PageData, chapterName string) (err error) {
	aditionalStyle := ""
	if page.Fill != "white" {
		aditionalStyle = "background-color:#000000;"
	}

	pathParts := strings.Split(payload.Path, string(filepath.Separator))
	relPath := filepath.Join("..", "..", "Images")
	for _, part := range pathParts[len(pathParts)-2:] {
		relPath += string(filepath.Separator) + part
	}

	htmlPath := filepath.Join(textPath, chapterName, payload.Title+".xhtml")

	deviceHeight := t.opts.ProfileData.Height
	imgWidth, imgHeight := (*payload.Image).Bounds().Dx(), (*payload.Image).Bounds().Dy()

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

	_, err = t.addFile(z, htmlPath, content)
	return
}

func getTopMargin(deviceHeight, imgHeight int) float64 {
	y := ((deviceHeight - imgHeight) / 2) / deviceHeight * 100
	return math.Round(float64(y*10)) / 10
}

func (t *EpubBuilder) BuildNCX(z *zip.Writer, uuid uuid.UUID) error {
	path := filepath.Join("OEBPS", "toc.ncx")

	content := fmt.Sprintf(
		EpubTemplates.NCXStart,
		uuid,
		t.dstFileName,
	)

	for _, chapter := range t.chapters {
		folder := filepath.Join(
			"Text",
			chapter.NormalizedName,
			chapter.Pages[0].Payloads[0].Title+".xhtml",
		)

		content += fmt.Sprintf(
			EpubTemplates.NCXNavPoint,
			strings.Replace(folder, string(filepath.Separator), "_", -1),
			chapter.Title,
			filepath.ToSlash(folder),
		)
	}

	content += EpubTemplates.NCXEnd

	_, err := t.addFile(z, path, content)
	return err
}

func (t *EpubBuilder) BuildNAV(z *zip.Writer) error {
	path := filepath.Join("OEBPS", "nav.xhtml")

	content := fmt.Sprintf(
		EpubTemplates.NAVStart,
		t.dstFileName,
	)

	var listContent string

	for _, chapter := range t.chapters {
		folder := filepath.Join(
			"Text",
			chapter.NormalizedName,
			chapter.Pages[0].Payloads[0].Title+".xhtml",
		)

		listContent += fmt.Sprintf(
			EpubTemplates.NAVLiElem,
			filepath.ToSlash(folder),
			chapter.Title,
		)
	}

	content += listContent +
		EpubTemplates.NAVBetweenList +
		listContent +
		EpubTemplates.NAVEnd

	_, err := t.addFile(z, path, content)
	return err
}

func (t *EpubBuilder) BuildOPF(z *zip.Writer, uuid uuid.UUID) error {
	path := filepath.Join("OEBPS", "content.opf")

	content := fmt.Sprintf(
		EpubTemplates.OPFStart,
		t.dstFileName,
		uuid,
		"0.0",
	)

	writingMode := "horizontal-lr"
	if t.opts.Manga {
		writingMode = "horizontal-rl"
	}

	content += fmt.Sprintf(
		EpubTemplates.OPFMetas,
		time.Now().UTC().Format(time.RFC3339),
		t.opts.ProfileData.Width,
		t.opts.ProfileData.Height,
		writingMode,
		false,
	)

	var refList []string
	for _, chapter := range t.chapters {
		for _, page := range chapter.Pages {
			for _, payload := range page.Payloads {
				folder := filepath.Join(
					chapter.NormalizedName,
					payload.Title,
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
	if t.opts.Manga {
		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "rtl")
		pageSide = "right"
	} else {
		content += fmt.Sprintf(EpubTemplates.OPFPageProgression, "ltr")
		pageSide = "left"
	}

	if t.opts.SpreadShift {
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
		case "-kcc-a", "-kcc-d":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "center")
			pageSide = calculatePageSide(t.opts.Manga)
		case "-kcc-b":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "right")
			pageSide = calculatePageSide(t.opts.Manga)
		case "-kcc-c":
			pageSpreadPropertyList = append(pageSpreadPropertyList, "left")
			pageSide = calculatePageSide(t.opts.Manga)
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

		if "-kcc-x" != ending {
			spreadSeen = true
			if t.opts.Manga {
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

	_, err := t.addFile(z, path, content)
	return err
}

func calculatePageSide(manga bool) string {
	if manga {
		return "right"
	}
	return "left"
}
