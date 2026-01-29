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

	isColor := pp.settings.ForceColor && processor.IsColored()
	if !isColor {
		processor.Grayscale()
	}

	if pp.settings.RemoveRainbowEffect {
		processor.RemoveRainbowEffect()
	}

	partProcessors := processor.TrySplit(pp.settings.SpreadSplitter == 2)
	if pp.settings.SpreadSplitter != 1 && len(partProcessors) > 2 {
		partProcessors = partProcessors[:2]
	}

	if len(partProcessors) >= 2 && pp.settings.RightToLeft {
		aux := partProcessors[1]
		partProcessors[1] = partProcessors[0]
		partProcessors[0] = aux
	}

	dir := filepath.Dir(page.Path)

	for _, partProcessor := range partProcessors {
		partProcessor.Resize()
		part := domain.NewPagePart(
			partProcessor.Img,
			partProcessor.SplitOperation,
		)

		if pp.settings.RightToLeft {
			switch partProcessor.SplitOperation {
			case image.ToLeft:
				partProcessor.SplitOperation = image.ToRight
			case image.ToRight:
				partProcessor.SplitOperation = image.ToLeft
			}
		}

		path := filepath.Join(dir, fmt.Sprintf("ermc-%d%c", num, part.PathOrder))
		path, err = partProcessor.SaveToDir(path)
		if err != nil {
			return err
		}
		part.SetPath(path)

		part.Clean()
		page.Parts = append(page.Parts, part)
	}

	return nil
}
