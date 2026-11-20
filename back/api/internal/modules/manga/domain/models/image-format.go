package models

type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
)

func (f ImageFormat) Ext() string {
	if f == ImageFormatPNG {
		return ".png"
	}
	return ".jpg"
}

func (f ImageFormat) MediaType() string {
	if f == ImageFormatPNG {
		return "image/png"
	}
	return "image/jpeg"
}
