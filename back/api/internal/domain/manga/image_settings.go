package manga

type ImageSettings struct {
	RightToLeft         bool
	SpreadSplitter      int // 0: Split, 1: Split+Rotated, 2: Rotated
	ForceColor          bool
	CroppingMode        int // 0: No, 1: Margins, 2: Margins + page numbers
	CroppingPower       float32
	RemoveRainbowEffect bool
	SetExtremBlackPoint bool
}

func NewDefaultImageSettings() *ImageSettings {
	return &ImageSettings{
		RightToLeft:         true,
		SpreadSplitter:      2,
		ForceColor:          true,
		CroppingMode:        2,
		CroppingPower:       1,
		RemoveRainbowEffect: true,
		SetExtremBlackPoint: true,
	}
}
