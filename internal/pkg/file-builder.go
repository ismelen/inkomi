package pkg

import (
	"archive/zip"
	"fmt"
	"strings"
)

type FileBuilder struct {
	parts []string
}

func NewFileBuilder() *FileBuilder { return &FileBuilder{} }

func (this *FileBuilder) AddFromTemplate(template string, values ...any) *FileBuilder {
	this.parts = append(this.parts, fmt.Sprintf(template, values...))
	return this
}

func (this *FileBuilder) Add(value string) *FileBuilder {
	this.parts = append(this.parts, value)
	return this
}

func (this *FileBuilder) BuildToZip(z *zip.Writer, path string) error {
	w, err := z.Create(path)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(this.Build()))
	if err != nil {
		return err
	}

	return nil
}

func (this *FileBuilder) Build() string {
	return strings.Join(this.parts, "")
}
