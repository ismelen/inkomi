package dtos

import (
	"encoding/json"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// ConfigDTO is both the manga config sent by the front and the mangaCfg the manga-worker reads.
type ConfigDTO struct {
	Hash                string                 `json:"-"`
	EReaderKey          string                 `json:"ereaderKey"`
	RotateDoublePage    bool                   `json:"rotateDoublePage"`
	CropTolerance       float32                `json:"cropTolerance"` // [0.0, 1.0] or -1 to disable; def=-1
	Enlarge             bool                   `json:"enlarge"`
	Embed               bool                   `json:"embed"`
	Gamma               float32                `json:"gamma"`      // [0.1, 3.0]; def=1.0
	Brightness          float32                `json:"brightness"` // [-1.0, 1.0]
	Contrast            float32                `json:"contrast"`   // [0.0, 2.0]; def=1.0
	Grayscale           bool                   `json:"grayscale"`
	BlackPointThreshold float32                `json:"blackPointThreshold"` // [0.0, 1.0]; def=0.06
	WhitePointThreshold float32                `json:"whitePointThreshold"` // [0.0, 1.0]; def=0.95
	Dithering           bool                   `json:"dithering"`
	Palette8bit         bool                   `json:"palette8bit"` // convert to 8-bit palette instead of 24-bit
	Quality             int8                   `json:"quality"`     // [1, 100]; def=90
	Sharpen             MangaSharpenOptionsDTO `json:"sharpen"`
	Kepubify            bool                   `json:"-"`
}

type MangaSharpenOptionsDTO struct {
	Enabled bool    `json:"enabled"`
	Radius  int8    `json:"radius"` // [1, 5]; def=1
	X1      float32 `json:"x1"`     // [1.0, 3.0]; def=1.5
	Y2      float32 `json:"y2"`     // [0.0, 50.0]; def=20
	Y3      float32 `json:"y3"`     // [0.0, 50.0]; def=50
	M1      float32 `json:"m1"`     // [0.0, +inf]; def=0
	M2      float32 `json:"m2"`     // [1.0, 5.0]; def=3
}

func (m *ConfigDTO) UnmarshalJSON(data []byte) error {
	*m = ConfigDTO{
		CropTolerance:       -1,
		Gamma:               1.0,
		Contrast:            1.0,
		BlackPointThreshold: 0.06,
		WhitePointThreshold: 0.95,
		Quality:             90,
		Sharpen: MangaSharpenOptionsDTO{
			Radius: 1,
			X1:     1.5,
			Y2:     20,
			Y3:     50,
			M2:     3,
		},
	}
	type Alias ConfigDTO
	aux := (*Alias)(m)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ereader, err := models.NewEreader(m.EReaderKey)
	if err != nil {
		return err
	}

	m.Kepubify = ereader.IsKepub
	return nil
}

func (m ConfigDTO) MarshalJSON() ([]byte, error) {
	ereader, err := models.NewEreader(m.EReaderKey)
	if err != nil {
		return nil, err
	}

	type Alias ConfigDTO

	return json.Marshal(&struct {
		Alias
		TargetWidth  int `json:"targetWidth"`
		TargetHeight int `json:"targetHeight"`
	}{
		Alias:        (Alias)(m),
		TargetWidth:  ereader.Width,
		TargetHeight: ereader.Height,
	})
}
