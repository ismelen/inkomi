package manga

import (
	"fmt"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/image"
	"path/filepath"
)

type PageProcessor struct {
	settings *domain.Settings
}

func NewPageProcessor(settings *domain.Settings) *PageProcessor {
	return &PageProcessor{settings: settings}
}

func (pp *PageProcessor) SetSettings(settings *domain.Settings) {
	pp.settings = settings
}

func (pp *PageProcessor) ProcessNewPage(page *domain.Page, num int) error {
	processor, err := image.NewProcessor(
		page.Path,
		pp.settings.Profile.Width,
		pp.settings.Profile.Height,
		pp.settings.ForceColor,
	)

	if err != nil {
		return err
	}

	page.HasWhiteBg = processor.HasWhiteBg()
	processor.CropMargins()

	isColor := pp.settings.ForceColor && processor.IsColored()
	if !isColor {
		processor.Grayscale()
	}

	if pp.settings.RemoveRainbowEffect && isColor {
		processor.RemoveRainbowEffect()
	}

	partPrcs := processor.TrySplit(pp.settings.SpreadSplitter == 2)
	
	if pp.settings.SpreadSplitter != 1 && len(partPrcs) > 2 {
		partPrcs = partPrcs[:2]
	}

	for _, partPrc := range partPrcs {
		partPrc.Resize()
		part := domain.NewPagePart(
			partPrc.Img,
			partPrc.SplitOperation,
		)

		dir := filepath.Dir(page.Path)
		path := filepath.Join(dir, fmt.Sprintf("ermc-%d%c", num, part.PathOrder))
		path, err = partPrc.SaveToDir(path)
		if err != nil {
			return err
		}
		part.SetPath(path)

		part.Clean()
		page.Parts = append(page.Parts, part)
	}

	return nil
}
