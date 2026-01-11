package manga

type ERProfile struct {
	Label   string
	Width   int
	Height  int
	Palette []uint8
	Gamma   float64
	IsKepub   bool
}
