package image

import (
	"fmt"
	"ismelen/inkomi/internal/domain/manga"
	"path/filepath"
)

type PageProcessor struct{}

func NewPageProcessor() *PageProcessor { return &PageProcessor{} }

func (p *PageProcessor) ProcessPage(path string, idx int, profile *manga.Profile, settings *manga.ImageSettings) (*manga.Page, error) {
	page := manga.NewPage(path)
	editor, err := NewEditor(
		path,
		profile.Width,
		profile.Height,
		settings.ForceColor,
	)
	if err != nil {
		return nil, err
	}

	page.HasWhiteBg = editor.HasWhiteBg()
	editor.CropMargins()

	isColor := settings.ForceColor && editor.IsColored()
	if !isColor {
		editor.Grayscale()
	}
	if settings.RemoveRainbowEffect && isColor {
		editor.RemoveRainbowEffect()
	}

	partEditors := editor.TrySplit(settings.SpreadSplitter == 2)
	if settings.SpreadSplitter != 1 && len(partEditors) > 2 {
		partEditors = partEditors[:2]
	}

	for _, partEditor := range partEditors {
		partEditor.Resize()
		part := manga.NewPagePart(
			partEditor.Img,
			partEditor.SplitOperation,
		)

		partPath := filepath.Join(
			filepath.Dir(path),
			fmt.Sprintf("inkomi-%d%c", idx, part.PathOrder),
		)
		partPath, err = partEditor.SaveToDir(partPath)
		if err != nil {
			return nil, err
		}

		part.SetPath(partPath)
		part.Clean()
		page.Parts = append(page.Parts, part)
	}

	return page, nil
}
