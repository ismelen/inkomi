package epub_test

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeEpub struct {
	title       string
	pages       int
	cover       bool
	fixedLayout bool
	// imagesOutsideOPF puts the OPF in OEBPS/ and the images in /Images, linked with "../".
	imagesOutsideOPF bool
}

// writeFakeEpub builds a minimal epub3 with one xhtml page (and image) per page.
func writeFakeEpub(t *testing.T, opts fakeEpub) string {
	t.Helper()

	epubPath := filepath.Join(t.TempDir(), "book.epub")
	file, err := os.Create(epubPath)
	require.NoError(t, err)
	zw := zip.NewWriter(file)

	add := func(name, content string) {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(content))
		require.NoError(t, err)
	}

	// imageHref is relative to the OPF, imageSrc to the xhtml pages in OEBPS/Text.
	imageDir, imageHref, imageSrc := "OEBPS/Images", "Images", "../Images"
	if opts.imagesOutsideOPF {
		imageDir, imageHref, imageSrc = "Images", "../Images", "../../Images"
	}

	var manifest, spine strings.Builder
	for i := range opts.pages {
		img := fmt.Sprintf("p%d.jpg", i)
		add(path.Join(imageDir, img), "jpeg")
		add(fmt.Sprintf("OEBPS/Text/p%d.xhtml", i), fmt.Sprintf(
			`<html xmlns="http://www.w3.org/1999/xhtml"><body><img src="%s/%s"/></body></html>`, imageSrc, img))
		fmt.Fprintf(&manifest, `<item id="img%d" href="%s/%s" media-type="image/jpeg"/>`, i, imageHref, img)
		fmt.Fprintf(&manifest, `<item id="page%d" href="Text/p%d.xhtml" media-type="application/xhtml+xml"/>`, i, i)
		spread := ""
		if opts.fixedLayout {
			spread = fmt.Sprintf(` properties="page-spread-%s"`, []string{"right", "left"}[i%2])
		}
		fmt.Fprintf(&spine, `<itemref idref="page%d"%s/>`, i, spread)
	}
	if opts.cover {
		add("OEBPS/Images/cover.jpg", "jpeg")
		manifest.WriteString(`<item id="cover" href="Images/cover.jpg" media-type="image/jpeg" properties="cover-image"/>`)
	}
	add("OEBPS/nav.xhtml", `<html xmlns="http://www.w3.org/1999/xhtml"><body><nav/></body></html>`)
	manifest.WriteString(`<item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>`)

	metas, prefix := "", ""
	if opts.fixedLayout {
		prefix = ` prefix="rendition: http://www.idpf.org/2013/rendition"`
		metas = `<meta property="rendition:layout">pre-paginated</meta>
<meta property="rendition:spread">landscape</meta>
<meta name="book-type" content="comic"/>
<meta name="primary-writing-mode" content="horizontal-lr"/>
<meta property="title-type" refines="#t">main</meta>`
	}

	add("META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`)
	add("OEBPS/content.opf", fmt.Sprintf(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id"%s>
<metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:identifier id="id">x</dc:identifier><dc:title>%s</dc:title><dc:language>ja</dc:language>%s
</metadata>
<manifest>%s</manifest>
<spine>%s</spine>
</package>`, prefix, opts.title, metas, manifest.String(), spine.String()))

	require.NoError(t, zw.Close())
	require.NoError(t, file.Close())
	return epubPath
}

type mergedOPF struct {
	Prefix   string `xml:"prefix,attr"`
	Metadata struct {
		Title    string `xml:"title"`
		Language string `xml:"language"`
		Metas    []struct {
			Name     string `xml:"name,attr"`
			Content  string `xml:"content,attr"`
			Property string `xml:"property,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest []struct {
		Id         string `xml:"id,attr"`
		Href       string `xml:"href,attr"`
		Properties string `xml:"properties,attr"`
	} `xml:"manifest>item"`
	Spine struct {
		Direction string `xml:"page-progression-direction,attr"`
		ItemRefs  []struct {
			IdRef      string `xml:"idref,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

type mergedEpub struct {
	files map[string]string
	order []string
	first *zip.FileHeader
	opf   mergedOPF
}

// readMerged opens an epub written by the merger and parses its OPF.
func readMerged(t *testing.T, epubPath string) mergedEpub {
	t.Helper()

	zr, err := zip.OpenReader(epubPath)
	require.NoError(t, err)
	defer zr.Close()

	out := mergedEpub{files: map[string]string{}, first: &zr.File[0].FileHeader}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		out.files[f.Name] = string(data)
		out.order = append(out.order, f.Name)
	}

	require.NoError(t, xml.Unmarshal([]byte(out.files["OEBPS/content.opf"]), &out.opf))
	return out
}

func (m mergedEpub) meta(key string) (string, bool) {
	for _, meta := range m.opf.Metadata.Metas {
		switch key {
		case meta.Name:
			return meta.Content, true
		case meta.Property:
			return meta.Value, true
		}
	}
	return "", false
}
