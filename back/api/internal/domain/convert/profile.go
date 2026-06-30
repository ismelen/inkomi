package convert

import "ismelen/inkomi/internal/domain/manga"

// NewProfile returns the manga.Profile for the given device label.
func NewProfile(label string) (*manga.Profile, error) {
	profile, ok := Profiles[label]
	if !ok {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

// Profiles maps device labels to their screen dimensions and format.
var Profiles = map[string]manga.Profile{
	"K1":    {Width: 600, Height: 670, IsKepub: false},
	"K2":    {Width: 600, Height: 670, IsKepub: false},
	"KDX":   {Width: 824, Height: 1000, IsKepub: false},
	"K34":   {Width: 600, Height: 800, IsKepub: false},
	"K57":   {Width: 600, Height: 800, IsKepub: false},
	"KPW":   {Width: 758, Height: 1024, IsKepub: false},
	"KV":    {Width: 1072, Height: 1448, IsKepub: false},
	"KPW34": {Width: 1072, Height: 1448, IsKepub: false},
	"K810":  {Width: 600, Height: 800, IsKepub: false},
	"KO":    {Width: 1264, Height: 1680, IsKepub: false},
	"K11":   {Width: 1072, Height: 1448, IsKepub: false},
	"KPW5":  {Width: 1236, Height: 1648, IsKepub: false},
	"KS":    {Width: 1860, Height: 2480, IsKepub: false},
	"KCS":   {Width: 1264, Height: 1680, IsKepub: false},

	// Kobo
	"KoMT":   {Width: 600, Height: 800, IsKepub: true},
	"KoG":    {Width: 768, Height: 1024, IsKepub: true},
	"KoGHD":  {Width: 1072, Height: 1448, IsKepub: true},
	"KoA":    {Width: 758, Height: 1024, IsKepub: true},
	"KoAHD":  {Width: 1080, Height: 1440, IsKepub: true},
	"KoAH2O": {Width: 1080, Height: 1430, IsKepub: true},
	"KoAO":   {Width: 1404, Height: 1872, IsKepub: true},
	"KoN":    {Width: 758, Height: 1024, IsKepub: true},
	"KoC":    {Width: 1072, Height: 1448, IsKepub: true},
	"KoCC":   {Width: 1072, Height: 1448, IsKepub: true},
	"KoL":    {Width: 1264, Height: 1680, IsKepub: true},
	"KoLC":   {Width: 1264, Height: 1680, IsKepub: true},
	"KoF":    {Width: 1440, Height: 1920, IsKepub: true},
	"KoS":    {Width: 1440, Height: 1920, IsKepub: true},
	"KoE":    {Width: 1404, Height: 1872, IsKepub: true},
}
