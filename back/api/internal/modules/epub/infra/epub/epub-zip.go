package epub

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
)

// epubWriter writes an epub zip: the stored mimetype first, then the other entries in call order.
type epubWriter struct {
	file    *os.File
	zw      *zip.Writer
	written map[string]bool
}

func newEpubWriter(outputPath string) (*epubWriter, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", outputPath, err)
	}

	w := &epubWriter{file: file, zw: zip.NewWriter(file), written: map[string]bool{}}
	mime, err := w.zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err == nil {
		_, err = mime.Write([]byte("application/epub+zip"))
	}
	if err != nil {
		return nil, errorAndClose(fmt.Errorf("write mimetype: %w", err), file)
	}
	return w, nil
}

// copy copies a source entry under name without recompressing it.
func (w *epubWriter) copy(f *zip.File, name string) error {
	if w.written[name] {
		return nil
	}
	w.written[name] = true

	header := f.FileHeader
	header.Name = name
	dst, err := w.zw.CreateRaw(&header)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	src, err := f.OpenRaw()
	if err != nil {
		return fmt.Errorf("open %s: %w", f.Name, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy %s: %w", f.Name, err)
	}
	return nil
}

func (w *epubWriter) create(name string) (io.Writer, error) {
	dst, err := w.zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", name, err)
	}
	return dst, nil
}

func (w *epubWriter) write(name, content string) error {
	dst, err := w.create(name)
	if err != nil {
		return err
	}
	if _, err := dst.Write([]byte(content)); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func (w *epubWriter) render(name, tmpl string, data any) error {
	dst, err := w.create(name)
	if err != nil {
		return err
	}
	if err := templates.ExecuteTemplate(dst, tmpl, data); err != nil {
		return fmt.Errorf("render %s: %w", name, err)
	}
	return nil
}

func (w *epubWriter) Close() error {
	return errors.Join(w.zw.Close(), w.file.Close())
}
