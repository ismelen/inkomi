package image

import FileUtils "ismelen/ermc/internal/utils/file"

func (ip *ImageProcessor) SaveToDir(path string) (string, error) {
	return FileUtils.SaveWithCodec(
		ip.Img,
		path,
	)
}
