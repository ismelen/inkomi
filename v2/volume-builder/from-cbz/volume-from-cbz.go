package fromCbz

import (
	"ismelen/ermc/v2/manga"
	"mime/multipart"
)

type VolumeFromCbz struct{}

func (v *VolumeFromCbz) FromMultipart(files ...*multipart.FileHeader) ([]*manga.Volume, error) {
	return nil, nil
}

func (v *VolumeFromCbz) FromPaths(paths ...string) ([]*manga.Volume, error) {
	return nil, nil
}
