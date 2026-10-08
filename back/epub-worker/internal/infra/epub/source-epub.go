package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
)

type epubContainer struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
	Manifest struct {
		Items []opfItem `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Toc      string       `xml:"toc,attr"`
		ItemRefs []opfItemRef `xml:"itemref"`
	} `xml:"spine"`
}

type opfMetadata struct {
	Title    string    `xml:"title"`
	Creator  string    `xml:"creator"`
	Language string    `xml:"language"`
	Metas    []opfMeta `xml:"meta"`
}

type opfMeta struct {
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
	Property string `xml:"property,attr"`
}

type opfItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type opfItemRef struct {
	IDRef  string `xml:"idref,attr"`
	Linear string `xml:"linear,attr"`
}

type sourceEpub struct {
	index    int
	baseName string
	opfDir   string
	opf      opfPackage
	files    map[string][]byte
}

func loadSourceEpub(epubPath string, index int) (*sourceEpub, error) {
	zr, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir %s: %w", epubPath, err)
	}
	defer zr.Close()

	files := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("%s: no se pudo leer %s: %w", epubPath, f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: error leyendo %s: %w", epubPath, f.Name, err)
		}
		files[f.Name] = data
	}

	opfRelPath, err := findOPFPath(files)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", epubPath, err)
	}
	opfData, ok := files[opfRelPath]
	if !ok {
		return nil, fmt.Errorf("%s: no se encontró el OPF en %q", epubPath, opfRelPath)
	}
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, fmt.Errorf("%s: OPF inválido: %w", epubPath, err)
	}

	return &sourceEpub{
		index:    index,
		baseName: strings.TrimSuffix(filepath.Base(epubPath), filepath.Ext(epubPath)),
		opfDir:   path.Dir(opfRelPath),
		opf:      pkg,
		files:    files,
	}, nil
}

func findOPFPath(files map[string][]byte) (string, error) {
	data, ok := files["META-INF/container.xml"]
	if !ok {
		return "", fmt.Errorf("falta META-INF/container.xml (¿no es un epub válido?)")
	}
	var c epubContainer
	if err := xml.Unmarshal(data, &c); err != nil {
		return "", fmt.Errorf("container.xml inválido: %w", err)
	}
	if len(c.Rootfiles) == 0 || c.Rootfiles[0].FullPath == "" {
		return "", fmt.Errorf("container.xml sin rootfile válido")
	}
	return c.Rootfiles[0].FullPath, nil
}

func detectCoverItemID(pkg opfPackage) string {
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(item.Properties, "cover-image") {
			return item.ID
		}
	}
	for _, m := range pkg.Metadata.Metas {
		if m.Name == "cover" && m.Content != "" {
			return m.Content
		}
	}
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(strings.ToLower(item.ID), "cover") && strings.HasPrefix(item.MediaType, "image/") {
			return item.ID
		}
	}
	return ""
}
