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

func (t *Options) ValidateAndNormalize() error {
	if err := t.normalizeInputs(); err != nil {
		return err
	}

	t.setTitle()
	t.setOutput()

	return nil
}

func (t *Options) setOutput() {
	if t.Output != "" {
		return
	}

	t.Output = SysUtils.NewTempDir("results")
}

func (t *Options) setTitle() {
	if t.Title != "" {
		return
	}

	t.Title = filepath.Base(t.Inputs[0])
	t.Title = strings.TrimSuffix(t.Title, filepath.Ext(t.Title))
}

func (t *Options) normalizeInputs() error {
	if len(t.Inputs) < 1 {
		return fmt.Errorf("No inputs")
	}

	var data []Utils.Pair[string, int64]

	if len(t.Inputs) > 1 {
		for _, path := range t.Inputs {
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
		ok, err := FileUtils.IsDir(t.Inputs[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("Not a directory")
		}

		data, err = SysUtils.GetChildsInfo(t.Inputs[0])
		if err != nil {
			return err
		}
	}

	for _, value := range data {
		t.InputData = append(
			t.InputData,
			NewChapterData(value.Fst, value.Snd),
		)
	}

	slices.SortFunc(t.InputData, func(a, b *ChapterData) int {
		return FileUtils.FilenameCmp(a.Path, b.Path)
	})

	t.InputData = t.InputData[:1]

	return nil
}
