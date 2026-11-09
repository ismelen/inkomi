package models

import "fmt"

type EReader struct {
	Key           string
	Width, Height int
	IsKepub       bool
}

func NewEreader(key string) (*EReader, error) {
	profile, ok := profiles[key]
	if !ok {
		return nil, fmt.Errorf("e-reader not available")
	}
	return &profile, nil
}

// profiles maps device labels to their screen dimensions and format.
var profiles = map[string]EReader{
	"K1":    {Key: "K1", Width: 600, Height: 670, IsKepub: false},
	"K2":    {Key: "K2", Width: 600, Height: 670, IsKepub: false},
	"KDX":   {Key: "KDX", Width: 824, Height: 1000, IsKepub: false},
	"K34":   {Key: "K34", Width: 600, Height: 800, IsKepub: false},
	"K57":   {Key: "K57", Width: 600, Height: 800, IsKepub: false},
	"KPW":   {Key: "KPW", Width: 758, Height: 1024, IsKepub: false},
	"KV":    {Key: "KV", Width: 1072, Height: 1448, IsKepub: false},
	"KPW34": {Key: "KPW34", Width: 1072, Height: 1448, IsKepub: false},
	"K810":  {Key: "K810", Width: 600, Height: 800, IsKepub: false},
	"KO":    {Key: "KO", Width: 1264, Height: 1680, IsKepub: false},
	"K11":   {Key: "K11", Width: 1072, Height: 1448, IsKepub: false},
	"KPW5":  {Key: "KPW5", Width: 1236, Height: 1648, IsKepub: false},
	"KS":    {Key: "KS", Width: 1860, Height: 2480, IsKepub: false},
	"KCS":   {Key: "KCS", Width: 1264, Height: 1680, IsKepub: false},

	// Kobo
	"KoMT":   {Key: "KoMT", Width: 600, Height: 800, IsKepub: true},
	"KoG":    {Key: "KoG", Width: 768, Height: 1024, IsKepub: true},
	"KoGHD":  {Key: "KoGHD", Width: 1072, Height: 1448, IsKepub: true},
	"KoA":    {Key: "KoA", Width: 758, Height: 1024, IsKepub: true},
	"KoAHD":  {Key: "KoAHD", Width: 1080, Height: 1440, IsKepub: true},
	"KoAH2O": {Key: "KoAH2O", Width: 1080, Height: 1430, IsKepub: true},
	"KoAO":   {Key: "KoAO", Width: 1404, Height: 1872, IsKepub: true},
	"KoN":    {Key: "KoN", Width: 758, Height: 1024, IsKepub: true},
	"KoC":    {Key: "KoC", Width: 1072, Height: 1448, IsKepub: true},
	"KoCC":   {Key: "KoCC", Width: 1072, Height: 1448, IsKepub: true},
	"KoL":    {Key: "KoL", Width: 1264, Height: 1680, IsKepub: true},
	"KoLC":   {Key: "KoLC", Width: 1264, Height: 1680, IsKepub: true},
	"KoF":    {Key: "KoF", Width: 1440, Height: 1920, IsKepub: true},
	"KoS":    {Key: "KoS", Width: 1440, Height: 1920, IsKepub: true},
	"KoE":    {Key: "KoE", Width: 1404, Height: 1872, IsKepub: true},
}
