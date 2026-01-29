package domain

import (
	"fmt"
	"ismelen/ermc/internal/pkg"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ImageSettings struct {
	RightToLeft         bool
	SpreadSplitter      int // 0: Split, 1: Split+Rotated, 2: Rotated
	ForceColor          bool
	CroppingMode        int // 0: No, 1: Margins, 2: Margins + page numbers
	CroppingPower       float32
	RemoveRainbowEffect bool
	SetExtremBlackPoint bool
}

func NewDefaultImageSettings() ImageSettings {
	return ImageSettings{
		RightToLeft:         true,
		SpreadSplitter:      2,
		ForceColor:          true,
		CroppingMode:        2,
		CroppingPower:       1,
		RemoveRainbowEffect: true,
		SetExtremBlackPoint: true,
	}
}

type Settings struct {
	ImageSettings

	Output         OutputPaths
	TargetSize     int64 // MB
	Author         string
	Title          string
	Profile        *Profile
	Merge          bool
	FirstVolumeNum int
	Volumes        []*Volume
}

type OutputPaths struct {
	Base     string
	Chapters string
}

func NewSettings(author, title, profile string, merge bool, FirstVolumeNum int) (*Settings, error) {
	eReaderProfile, ok := Profiles[profile]
	if !ok {
		var supportedProfiles []string
		for k := range Profiles {
			supportedProfiles = append(supportedProfiles, k)
		}
		return nil, echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("Profile (%s) not supported. Supported profiles: %s", profile, supportedProfiles),
		)
	}

	output, err := pkg.NewTempDir("ermc")
	if err != nil {
		return nil, err
	}

	var targetSize int64 = 200
	if !merge {
		targetSize = 0
	}

	if title == "" {
		title = pkg.NormalizeString(uuid.NewString())
	}

	return &Settings{
		Author:     author,
		Title:      title,
		Profile:    &eReaderProfile,
		TargetSize: targetSize,
		Output: OutputPaths{
			Base:     output,
			Chapters: filepath.Join(output, "chapters"),
		},
	}, nil
}

func (s *Settings) SetImageSettings(imageSettings ImageSettings) {
	s.ImageSettings = imageSettings
}

func (s *Settings) SetVolumes(volumes []*Volume) {
	s.Volumes = volumes
}

func (s *Settings) GetPageProgression() string {
	if s.RightToLeft {
		return "rtl"
	}
	return "ltr"
}

func (b *Settings) GetWritingMode() string {
	if b.RightToLeft {
		return "horizontal-rl"
	}
	return "horizontal-lr"
}
