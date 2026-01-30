package cloud

import "path/filepath"

type LocalCloud struct{}

func (*LocalCloud) Init(accesToken, folder string) error { return nil }
func (*LocalCloud) Upload(path string) (string, error) {
	filename := filepath.Base(path)
	dir := filepath.Dir(path)
	dir = filepath.Base(dir)
	return filepath.ToSlash(filepath.Join(dir, filename)), nil
}
