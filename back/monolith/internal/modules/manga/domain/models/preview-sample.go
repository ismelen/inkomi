package models

type PreviewSample string

const (
	PreviewSampleColor PreviewSample = "color"
	PreviewSampleBW    PreviewSample = "bw"
)

// PreviewMaxSize caps the target size of a preview, which is public and served on demand.
const PreviewMaxSize = 2000

func NewPreviewSample(value string) (PreviewSample, error) {
	switch sample := PreviewSample(value); sample {
	case PreviewSampleColor, PreviewSampleBW:
		return sample, nil
	default:
		return "", ErrUnknownSample
	}
}
