package manga

// ImageProcessor is the port implemented by infra/image to process a raw page file
// into a domain Page ready for the epub builder.
type ImageProcessor interface {
	ProcessPage(path string, idx int, profile *Profile, settings *ImageSettings) (*Page, error)
}
