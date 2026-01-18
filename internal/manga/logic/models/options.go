package manga

import (
	"fmt"
	Utils "ismelen/ermc/internal/utils"
	FileUtils "ismelen/ermc/internal/utils/file"
	SysUtils "ismelen/ermc/internal/utils/sys"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Options holds all the configuration for the conversion process.
type Options struct {
	Inputs     []string
	Output     string
	Profile    string
	Title      string
	Author     string
	Format     string // "Auto", "MOBI", "EPUB", "CBZ", "PDF", etc.
	TargetSize int64  // MB
	LowRAM     bool   // Low memory usage mode (slower)

	Manga              bool // Right-to-left
	SpreadShift        bool
	FileFusion         bool
	SpreadSplitter     int // 0: Split, 1: Split+Rotated, 2: Rotated
	ColorMode          bool
	CroppingMode       int // 0: No, 1: Margins, 2: Margins + page numbers
	CroppingPower      float32
	RainbowEraser      bool
	ExtremBlackPoint   bool
	StretchUpscaleMode int // 0: Nothing, 1: Stretching, 2: Upscaling
	PreserveMargin     float64

	// Internal calculated fields
	ProfileData ERProfile
	InputData   []*ChapterData
}

func NewOptions(input, profile, title, author string, merge bool) Options {
	return Options{
		Inputs: []string{input},
		Profile: profile,
		Title: title,
		Author: author,
		Format: "EPUB",
		TargetSize: 200,
		LowRAM: true,

		Manga: true,
		SpreadShift: true,
		FileFusion: merge,
		SpreadSplitter: 2,
		ColorMode: true,
		CroppingMode: 2,
		CroppingPower: 1.0,
		RainbowEraser: true,
		ExtremBlackPoint: true,
		StretchUpscaleMode: 2,
		PreserveMargin: 0.0,
	}
}

func (this *Options) ValidateAndNormalize() error {
	if err := this.setProfileData(); err != nil { return err }
	if err := this.normalizeInputs(); err != nil { return err }

	this.setTitle()
	this.setOutput()

	if this.FileFusion {
		this.TargetSize = this.TargetSize << 20
	}else{
		this.TargetSize = 0
	}

	return nil
}

func (this *Options) setProfileData() error {
	profileData, ok := Profiles[this.Profile]
	if !ok {
		return fmt.Errorf("Unknown profile: %s", this.Profile)
	}
	this.ProfileData = profileData
	return nil
}

func (this *Options) setOutput() {
	if this.Output != "" {
		return
	}

	this.Output = SysUtils.NewTempDir("results")
}

func (this *Options) setTitle() {
	if this.Title != "" {
		return
	}

	this.Title = filepath.Base(this.Inputs[0])
	this.Title = strings.TrimSuffix(this.Title, filepath.Ext(this.Title))
}

func (this *Options) normalizeInputs() error {
	if len(this.Inputs) < 1 {
		return fmt.Errorf("No inputs")
	}

	var data []Utils.Pair[string, int64]

	if len(this.Inputs) > 1 {
		for _, path := range this.Inputs {
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			data = append(data, Utils.Pair[string, int64]{
				Fst: path,
				Snd: info.Size(),
			})
		}
	} else {
		ok, err := FileUtils.IsDir(this.Inputs[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("Not a directory")
		}

		data, err = SysUtils.GetChildsInfo(this.Inputs[0])
		if err != nil {
			return err
		}
	}

	for _, value := range data {
		this.InputData = append(
			this.InputData,
			NewChapterData(value.Fst, value.Snd),
		)
	}

	slices.SortFunc(this.InputData, func(a, b *ChapterData) int {
		return FileUtils.FilenameCmp(a.Path, b.Path)
	})

	return nil
}
