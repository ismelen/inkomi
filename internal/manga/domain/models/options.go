package MangaModels

import (
	"fmt"
	Utils "ismelen/ermc/internal/utils"
	FileUtils "ismelen/ermc/internal/utils/file"
	StringUtils "ismelen/ermc/internal/utils/strings"
	SysUtils "ismelen/ermc/internal/utils/sys"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ConverterOptions holds all the configuration for the conversion process.
type ConverterOptions struct {
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
	InputData   []*Chapter
}

func NewOptions(input, profile, title, author string, merge bool) ConverterOptions {
	return ConverterOptions{
		Inputs: []string{input},
		Profile: profile,
		Title: title,
		Author: author,
		Format: "EPUB",
		TargetSize: 200,
		LowRAM: false,

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

func (this *ConverterOptions) ValidateAndNormalize() error {
	if err := this.setProfileData(); err != nil { return err }

	this.setTitle()
	this.setOutput()

	if this.FileFusion {
		this.TargetSize = this.TargetSize << 20
	}else{
		this.TargetSize = 0
	}

	return nil
}

func (this *ConverterOptions) setProfileData() error {
	profileData, ok := Profiles[this.Profile]
	if !ok {
		return fmt.Errorf("Unknown profile: %s", this.Profile)
	}
	this.ProfileData = profileData
	return nil
}

func (this *ConverterOptions) setOutput() {
	if this.Output != "" {
		return
	}

	this.Output = SysUtils.NewTempDir("results")
}

func (this *ConverterOptions) setTitle() {
	if this.Title != "" {
		return
	}

	this.Title = filepath.Base(this.Inputs[0])
	this.Title = strings.TrimSuffix(this.Title, filepath.Ext(this.Title))
	this.Title = StringUtils.NormalizeString(this.Title)
}

func (this *ConverterOptions) GetVolumes() ([]Volume, int, error) {
	hasInputs := len(this.Inputs) > 0
	if !hasInputs { return nil, 0, fmt.Errorf("No inputs") }

	var chaptersMetadata []Utils.Pair[string, int64]

	
	for _, path := range this.Inputs {
		metadata, err := this.getChaptersMetadata(path)
		if err != nil { return nil, 0, err }
		
		chaptersMetadata = append(chaptersMetadata, metadata...)
	}
	
	slices.SortFunc(chaptersMetadata, func(a, b Utils.Pair[string, int64]) int {
		return FileUtils.FilenameCmp(a.Fst, b.Fst)
	})

	var volumes []Volume

	if !this.FileFusion {
		for _, metadata := range chaptersMetadata {
			chapter := NewChapter(metadata.Fst)
			volumes = append(volumes, NewVolume(
				chapter.NormalizedName, 
				[]*Chapter{chapter},
			))
		}
		return volumes, len(chaptersMetadata), nil
	}

	var size int64 = 0
	var volIdx int
	chaptersCant := len(chaptersMetadata)
	var chapters []*Chapter

	for idx, metadata := range chaptersMetadata {
		size += metadata.Snd
		isLast := idx >= chaptersCant-1
		chapters = append(chapters, NewChapter(metadata.Fst))
		
		if size < this.TargetSize && !isLast { continue }

		volumes = append(volumes, NewVolume(
			fmt.Sprintf("%s Vol_%d", this.Title, volIdx+1),
			chapters,
		))

		chapters = []*Chapter{}
		volIdx++
		size = 0
	}
	
	return volumes , len(chaptersMetadata),nil
}

func (this *ConverterOptions) getChaptersMetadata(path string) ([]Utils.Pair[string, int64], error) {
	isDir, err := FileUtils.IsDir(path)
	if err != nil { return nil, err }

	var data []Utils.Pair[string, int64]

	if isDir {
		data, err = SysUtils.GetChildsInfo(path)
		if err != nil { return nil, err }
	} else{
		info, err := os.Stat(path)
		if err != nil { return nil, err }
		data = append(data, Utils.Pair[string, int64]{
			Fst: path,
			Snd: info.Size(),
		})
	}

	return data, nil
}

func (this *ConverterOptions) GetWritingMode() string {
	if this.Manga {
		return "horizontal-rl" 
	}
	return "horizontal-lr"
}

func (this *ConverterOptions) GetPageProgression() string {
	if this.Manga {
		return "rtl"
	}
	return "ltr"
}

func (this *ConverterOptions) GetPageSide() string {
	if this.Manga {
		return "right"
	}
	return "left"
}

func (this *ConverterOptions) GetSpreadShiftPageSide() string {
	side := this.GetPageSide()
	if this.SpreadShift {
		return StringUtils.Toggle(side, "right", "left")
	}
	return side
}
