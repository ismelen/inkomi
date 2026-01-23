package ContentBuilder

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"
)

type ContentBuilder struct {
	parts []string
}

func New() *ContentBuilder { return &ContentBuilder{} }

func (this *ContentBuilder) AddFromTemplate(template string, values ...any) *ContentBuilder {
	this.parts = append(this.parts, fmt.Sprintf(template, values...))
	return this
}

func (this *ContentBuilder) Add(value string) *ContentBuilder {
	this.parts = append(this.parts, value)
	return this
}

func (this *ContentBuilder) BuildToZip(z *zip.Writer, path string) error {
	w, err := z.Create(filepath.ToSlash(path))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(this.Build()))
	if err != nil {
		return err
	}

	return nil
}

func (this *ContentBuilder) Build() string {
	return strings.Join(this.parts, "")
}
