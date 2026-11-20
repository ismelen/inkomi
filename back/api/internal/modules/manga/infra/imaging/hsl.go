package imaging

import "math"

type hsl struct {
	h, s, l float64
}

func rgbToHSL(r, g, b float64) hsl {
	hi := math.Max(math.Max(r, g), b)
	lo := math.Min(math.Min(r, g), b)
	l := (hi + lo) / 2
	if hi == lo {
		return hsl{0, 0, l}
	}

	d := hi - lo
	var s float64
	if l > 0.5 {
		s = d / (2 - hi - lo)
	} else {
		s = d / (hi + lo)
	}

	var h float64
	switch hi {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return hsl{h / 6, s, l}
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	return hueToRGB(p, q, h+1.0/3), hueToRGB(p, q, h), hueToRGB(p, q, h-1.0/3)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}
