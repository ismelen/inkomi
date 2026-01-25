package image

func (ip *ImageProcessor) OptimizeForDisplay() {
	bounds := (*ip.Img).Bounds()
	if bounds.Dx() > 1 && bounds.Dy() > 1 {
		ip.RemoveRainbowEffect()
	}
}
