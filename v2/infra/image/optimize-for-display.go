package image

func (ip *ImageEditor) OptimizeForDisplay() {
	bounds := (*ip.Img).Bounds()
	if bounds.Dx() > 1 && bounds.Dy() > 1 {
		ip.RemoveRainbowEffect()
	}
}
