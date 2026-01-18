package manga

type ERProfile struct {
	Label   string
	Width   int
	Height  int
	Palette []uint8
	Gamma   float64
	IsKepub bool
}

var Profiles = map[string]ERProfile{
	"K1":    {"Kindle 1", 600, 670, Palette4, 1.0, false},
	"K2":    {"Kindle 2", 600, 670, Palette15, 1.0, false},
	"KDX":   {"Kindle DX/DXG", 824, 1000, Palette16, 1.0, false},
	"K34":   {"Kindle Keyboard/Touch", 600, 800, Palette16, 1.0, false},
	"K57":   {"Kindle 5/7", 600, 800, Palette16, 1.0, false},
	"KPW":   {"Kindle Paperwhite 1/2", 758, 1024, Palette16, 1.0, false},
	"KV":    {"Kindle Voyage", 1072, 1448, Palette16, 1.0, false},
	"KPW34": {"Kindle Paperwhite 3/4/Oasis", 1072, 1448, Palette16, 1.0, false},
	"K810":  {"Kindle 8/10", 600, 800, Palette16, 1.0, false},
	"KO":    {"Kindle Oasis 2/3/Paperwhite 12", 1264, 1680, Palette16, 1.0, false},
	"K11":   {"Kindle 11", 1072, 1448, Palette16, 1.0, false},
	"KPW5":  {"Kindle Paperwhite 5/Signature Edition", 1236, 1648, Palette16, 1.0, false},
	"KS":    {"Kindle Scribe", 1860, 2480, Palette16, 1.0, false},
	"KCS":   {"Kindle Colorsoft", 1264, 1680, Palette16, 1.0, false},

	// Kobo
	"KoMT":   {"Kobo Mini/Touch", 600, 800, Palette16, 1.0, true},
	"KoG":    {"Kobo Glo", 768, 1024, Palette16, 1.0, true},
	"KoGHD":  {"Kobo Glo HD", 1072, 1448, Palette16, 1.0, true},
	"KoA":    {"Kobo Aura", 758, 1024, Palette16, 1.0, true},
	"KoAHD":  {"Kobo Aura HD", 1080, 1440, Palette16, 1.0, true},
	"KoAH2O": {"Kobo Aura H2O", 1080, 1430, Palette16, 1.0, true},
	"KoAO":   {"Kobo Aura ONE", 1404, 1872, Palette16, 1.0, true},
	"KoN":    {"Kobo Nia", 758, 1024, Palette16, 1.0, true},
	"KoC":    {"Kobo Clara HD/Kobo Clara 2E", 1072, 1448, Palette16, 1.0, true},
	"KoCC":   {"Kobo Clara Colour", 1072, 1448, Palette16, 1.0, true},
	"KoL":    {"Kobo Libra H2O/Kobo Libra 2", 1264, 1680, Palette16, 1.0, true},
	"KoLC":   {"Kobo Libra Colour", 1264, 1680, Palette16, 1.0, true},
	"KoF":    {"Kobo Forma", 1440, 1920, Palette16, 1.0, true},
	"KoS":    {"Kobo Sage", 1440, 1920, Palette16, 1.0, true},
	"KoE":    {"Kobo Elipsa", 1404, 1872, Palette16, 1.0, true},
}

// Color Palettes
var (
	Palette4 = []uint8{
		0x00, 0x00, 0x00,
		0x55, 0x55, 0x55,
		0xaa, 0xaa, 0xaa,
		0xff, 0xff, 0xff,
	}
	Palette15 = []uint8{
		0x00, 0x00, 0x00, 0x11, 0x11, 0x11, 0x22, 0x22, 0x22,
		0x33, 0x33, 0x33, 0x44, 0x44, 0x44, 0x55, 0x55, 0x55,
		0x66, 0x66, 0x66, 0x77, 0x77, 0x77, 0x88, 0x88, 0x88,
		0x99, 0x99, 0x99, 0xaa, 0xaa, 0xaa, 0xbb, 0xbb, 0xbb,
		0xcc, 0xcc, 0xcc, 0xdd, 0xdd, 0xdd, 0xff, 0xff, 0xff,
	}
	Palette16 = []uint8{
		0x00, 0x00, 0x00, 0x11, 0x11, 0x11, 0x22, 0x22, 0x22,
		0x33, 0x33, 0x33, 0x44, 0x44, 0x44, 0x55, 0x55, 0x55,
		0x66, 0x66, 0x66, 0x77, 0x77, 0x77, 0x88, 0x88, 0x88,
		0x99, 0x99, 0x99, 0xaa, 0xaa, 0xaa, 0xbb, 0xbb, 0xbb,
		0xcc, 0xcc, 0xcc, 0xdd, 0xdd, 0xdd, 0xee, 0xee, 0xee,
		0xff, 0xff, 0xff,
	}
)
