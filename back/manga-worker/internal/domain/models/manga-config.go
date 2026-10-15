package models

import (
	"encoding/json"
	"fmt"
	"math"
)

// MangaConfig is the image processing config of a manga job (jobs.Manga.MangaCfg).
type MangaConfig struct {
	RotateDoublePage    bool           `json:"rotateDoublePage"`
	TargetWidth         int            `json:"targetWidth"`
	TargetHeight        int            `json:"targetHeight"`
	CropTolerance       float64        `json:"cropTolerance"` // [0.0, 1.0] or -1 to disable cropping
	Enlarge             bool           `json:"enlarge"`       // upscale small images to the target size
	Embed               bool           `json:"embed"`         // fill blanks to fit the target size
	Gamma               float64        `json:"gamma"`
	Brightness          float64        `json:"brightness"`
	Contrast            float64        `json:"contrast"`
	Grayscale           bool           `json:"grayscale"`
	BlackPointThreshold float64        `json:"blackPointThreshold"`
	WhitePointThreshold float64        `json:"whitePointThreshold"`
	Dithering           bool           `json:"dithering"`
	Palette8bit         bool           `json:"palette8bit"` // 8-bit palette PNG instead of 24-bit JPEG
	Quality             int            `json:"quality"`
	Sharpen             SharpenOptions `json:"sharpen"`
}

type SharpenOptions struct {
	Enabled bool    `json:"enabled"`
	Radius  int     `json:"radius"`
	X1      float64 `json:"x1"`
	Y2      float64 `json:"y2"`
	Y3      float64 `json:"y3"`
	M1      float64 `json:"m1"`
	M2      float64 `json:"m2"`
}

const CropDisabled = -1

func DefaultMangaConfig() MangaConfig {
	return MangaConfig{
		TargetWidth:         1072,
		TargetHeight:        1448,
		CropTolerance:       CropDisabled,
		Gamma:               1.0,
		Contrast:            1.0,
		BlackPointThreshold: 0.06,
		WhitePointThreshold: 0.95,
		Quality:             90,
		Sharpen: SharpenOptions{
			Radius: 1,
			X1:     1.5,
			Y2:     20,
			Y3:     50,
			M1:     0,
			M2:     3,
		},
	}
}

// NewMangaConfig parses raw over the defaults, so missing fields keep their default value.
// An empty raw returns the defaults.
func NewMangaConfig(raw json.RawMessage) (MangaConfig, error) {
	cfg := DefaultMangaConfig()
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return MangaConfig{}, fmt.Errorf("%w: %w", ErrInvalidMangaConfig, err)
		}
	}
	if err := cfg.validate(); err != nil {
		return MangaConfig{}, fmt.Errorf("%w: %w", ErrInvalidMangaConfig, err)
	}
	return cfg, nil
}

// FitsPreview reports ErrPreviewTooLarge when the target size exceeds PreviewMaxSize.
func (c *MangaConfig) FitsPreview() error {
	if c.TargetWidth > PreviewMaxSize || c.TargetHeight > PreviewMaxSize {
		return fmt.Errorf("%w: max %dx%d", ErrPreviewTooLarge, PreviewMaxSize, PreviewMaxSize)
	}
	return nil
}

func (c *MangaConfig) CropEnabled() bool {
	return c.CropTolerance >= 0
}

type rangeCheck struct {
	name     string
	value    float64
	min, max float64
}

func (c *MangaConfig) validate() error {
	if !c.CropEnabled() && c.CropTolerance != CropDisabled {
		return fmt.Errorf("cropTolerance %v must be in [0, 1] or -1", c.CropTolerance)
	}

	checks := []rangeCheck{
		{"targetWidth", float64(c.TargetWidth), 1, 10000},
		{"targetHeight", float64(c.TargetHeight), 1, 10000},
		{"gamma", c.Gamma, 0.1, 3},
		{"brightness", c.Brightness, -1, 1},
		{"contrast", c.Contrast, 0, 2},
		{"blackPointThreshold", c.BlackPointThreshold, 0, 1},
		{"whitePointThreshold", c.WhitePointThreshold, 0, 1},
		{"quality", float64(c.Quality), 1, 100},
	}
	if c.CropEnabled() {
		checks = append(checks, rangeCheck{"cropTolerance", c.CropTolerance, 0, 1})
	}
	if c.Sharpen.Enabled {
		checks = append(checks,
			rangeCheck{"sharpen.radius", float64(c.Sharpen.Radius), 1, 5},
			rangeCheck{"sharpen.x1", c.Sharpen.X1, 1, 3},
			rangeCheck{"sharpen.y2", c.Sharpen.Y2, 0, 50},
			rangeCheck{"sharpen.y3", c.Sharpen.Y3, 0, 50},
			rangeCheck{"sharpen.m1", c.Sharpen.M1, 0, math.MaxFloat64},
			rangeCheck{"sharpen.m2", c.Sharpen.M2, 1, 5},
		)
	}

	for _, check := range checks {
		if check.value < check.min || check.value > check.max {
			return fmt.Errorf("%s %v out of range [%v, %v]", check.name, check.value, check.min, check.max)
		}
	}
	if c.BlackPointThreshold >= c.WhitePointThreshold {
		return fmt.Errorf("blackPointThreshold %v must be lower than whitePointThreshold %v", c.BlackPointThreshold, c.WhitePointThreshold)
	}
	return nil
}
