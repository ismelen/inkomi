package manga

import (
	"fmt"
	"ismelen/ermc/v2/image"
	"path/filepath"
)

type PageProcessor struct {
	settings *Settings
}

func NewPageProcessor(settings *Settings) *PageProcessor {
	return &PageProcessor{settings: settings}
}

func (pp *PageProcessor) SetSettings(settings *Settings) {
	pp.settings = settings
}

func (pp *PageProcessor) ProcessNewPage(page *Page, num int) error {
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

	processor.Resize()

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
		part := NewPagePart(
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

		path := filepath.Join(dir, fmt.Sprintf("ermc-%d%c", num, part.pathOrder))
		part.path, err = partProcessor.SaveToDir(path)
		if err != nil {
			return err
		}

		page.Parts = append(page.Parts, part)
	}

	return nil
}
