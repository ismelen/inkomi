package manga

import (
	documentBuilder "ismelen/ermc/v2/document-builder"
)

type Settings struct {
	Output            string
	LowRAM            bool
	TargetSize        int // MB
	Author            string
	Title             string
	DocumentProcessor documentBuilder.BuilderI
	Profile           *Profile

	RightToLeft         bool
	SpreadShift         bool
	FileFusion          bool
	SpreadSplitter      int // 0: Split, 1: Split+Rotated, 2: Rotated
	ForceColor          bool
	CroppingMode        int // 0: No, 1: Margins, 2: Margins + page numbers
	CroppingPower       float32
	RemoveRainbowEffect bool
	SetExtremBlackPoint bool
	// StretchUpscaleMode  int // 0: Nothing, 1: Stretching, 2: Upscaling
	PreserveMargin float64

	Volumes []*Volume
}

type Profile struct {
	Width, Height int
	IsKepub       bool
}
