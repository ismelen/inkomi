package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
)

type epubContainer struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opfPackage struct {
	Prefix   string      `xml:"prefix,attr"`
	Metadata opfMetadata `xml:"metadata"`
	Manifest struct {
		Items []opfItem `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		ItemRefs []opfItemRef `xml:"itemref"`
	} `xml:"spine"`
}

type opfMetadata struct {
	Title    string    `xml:"title"`
	Language string    `xml:"language"`
	Metas    []opfMeta `xml:"meta"`
}

type opfMeta struct {
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
	Property string `xml:"property,attr"`
	Refines  string `xml:"refines,attr"`
	Value    string `xml:",chardata"`
}

type opfItem struct {
	Id         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type opfItemRef struct {
	IdRef      string `xml:"idref,attr"`
	Linear     string `xml:"linear,attr"`
	Properties string `xml:"properties,attr"`
}

// sourceEpub is an opened epub whose entries are read lazily from the zip.
type sourceEpub struct {
	zip    *zip.ReadCloser
	files  map[string]*zip.File
	opfDir string
	opf    opfPackage
}

func openSource(epubPath string) (*sourceEpub, error) {
	zr, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", epubPath, err)
	}

	src := &sourceEpub{zip: zr, files: make(map[string]*zip.File, len(zr.File))}
	for _, f := range zr.File {
		src.files[f.Name] = f
	}

	if err := src.loadOPF(); err != nil {
		return nil, errorAndClose(fmt.Errorf("%s: %w", epubPath, err), zr)
	}
	return src, nil
}

func (s *sourceEpub) Close() error {
	return s.zip.Close()
}

func (s *sourceEpub) loadOPF() error {
	var container epubContainer
	if err := s.readXML("META-INF/container.xml", &container); err != nil {
		return err
	}
	if len(container.Rootfiles) == 0 || container.Rootfiles[0].FullPath == "" {
		return fmt.Errorf("container.xml without rootfile")
	}

	opfPath := container.Rootfiles[0].FullPath
	s.opfDir = path.Dir(opfPath)
	return s.readXML(opfPath, &s.opf)
}

func (s *sourceEpub) readXML(name string, v any) error {
	f, ok := s.files[name]
	if !ok {
		return fmt.Errorf("missing %s", name)
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if err := xml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

// entry returns the zip entry an OPF href points to.
func (s *sourceEpub) entry(href string) (*zip.File, bool) {
	unescaped, err := url.PathUnescape(href)
	if err != nil {
		unescaped = href
	}
	f, ok := s.files[path.Join(s.opfDir, unescaped)]
	return f, ok
}

// coverItemId finds the cover image declared by EPUB3 properties, the EPUB2 meta, or the item id.
func (s *sourceEpub) coverItemId() string {
	for _, item := range s.opf.Manifest.Items {
		if hasToken(item.Properties, "cover-image") {
			return item.Id
		}
	}
	for _, m := range s.opf.Metadata.Metas {
		if m.Name == "cover" && m.Content != "" {
			return m.Content
		}
	}
	for _, item := range s.opf.Manifest.Items {
		if strings.Contains(strings.ToLower(item.Id), "cover") && strings.HasPrefix(item.MediaType, "image/") {
			return item.Id
		}
	}
	return ""
}
