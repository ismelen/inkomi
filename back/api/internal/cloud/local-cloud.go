package cloud

import (
	"fmt"
	"path/filepath"
)

type LocalCloud struct{}

func (*LocalCloud) Init(accesToken, folder string) error { return nil }
func (*LocalCloud) Upload(path string) (string, error) {
	filename := filepath.Base(path)
	dir := filepath.Dir(path)
	fmt.Println(dir)
	dir = filepath.Base(dir)
	fmt.Println(dir)
	fmt.Println(filename)
	return filepath.ToSlash(filepath.Join(dir, filename)), nil
}
