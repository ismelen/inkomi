package dtos

import (
	"encoding/json"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type ConfigDTO struct {
	Hash                string
	EReaderKey          string                  `json:"ereader_key"`
	RotateDoublePage    bool                    `json:"rotate_double_page"`
	CropTolerance       float32                 `json:"crop_tolerance,omitempty"` // [0.0, 1.0]; def=0.0
	Enlarge             bool                    `json:"enlarge"`
	Embed               bool                    `json:"embed"`
	Gamma               float32                 `json:"gamma"`      // [0.1, 3.0] or [1.8, 2.6]; def=1.0
	Brightness          float32                 `json:"brightness"` // [-1.0, 1.0]
	Contrast            float32                 `json:"contrast"`   // [0.0, 2.0]
	Grayscale           bool                    `json:"grayscale"`
	BlackPointThreshold float32                 `json:"black_point_threshold"` // [0.0, 1.0], def=0.06
	WhitePointThreshold float32                 `json:"white_point_threshold"` // [0.0, 1.0], def=0.95
	Dithering           bool                    `json:"dithering"`
	Palette8bit         bool                    `json:"pallete_8_bit"` // convert to 8-bit palette instead of 24-bit
	Quality             int8                    `json:"quality"`       // %
	Sharpen             *MangaSharpenOptionsDTO `json:"sharpen,omitempty"`
	Kepubify            bool
}

type MangaSharpenOptionsDTO struct {
	Radius int8    `json:"radius"` // [1, 5]; def=1
	X1     float32 `json:"x1"`     // [1.0, 3.0]; def=1.5
	Y2     float32 `json:"y2"`     // [0.0, 50.0]; def=20
	Y3     float32 `json:"y3"`     // [0.0, 50.0]; def=50
	M1     float32 `json:"m1"`     // [0.0, +inf]; def=0
	M2     float32 `json:"m2"`     // [1.0, 5.0]; def=3
}

func (m *ConfigDTO) UnmarshalJSON(data []byte) error {
	*m = ConfigDTO{
		Gamma:               1.0,
		Contrast:            1.0,
		BlackPointThreshold: 0.06,
		WhitePointThreshold: 0.95,
	}
	type Alias ConfigDTO
	aux := (*Alias)(m)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Sharpen != nil && *aux.Sharpen == (MangaSharpenOptionsDTO{}) {
		m.Sharpen = nil
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
		TargetWidth  int `json:"target_width"`
		TargetHeight int `json:"target_height"`
	}{
		Alias:        (Alias)(m),
		TargetWidth:  ereader.Width,
		TargetHeight: ereader.Height,
	})

}
