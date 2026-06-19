package usecases

import (
	"context"
	"fmt"
	"ismelen/ermc/v2/domain"
	epubBuilder "ismelen/ermc/v2/infra/builders/epub-builder"
	"ismelen/ermc/v2/infra/cloud"
	"ismelen/ermc/v2/infra/image"
	"ismelen/ermc/v2/infra/state"
	"ismelen/ermc/v2/ports"
	"path/filepath"
)

type ConvertMangaUC struct {
	config        *domain.ConvertConfig
	profile       *domain.Profile
	imageSettings *domain.ImageSettings
	pushNotifier  ports.PushNotifier
}

func NewConvertMangaUC(pushNotifier ports.PushNotifier) *ConvertMangaUC {
	return &ConvertMangaUC{
		imageSettings: domain.NewDefaultImageSettings(),
		pushNotifier:  pushNotifier,
	}
}

func (c *ConvertMangaUC) Execute(chapters []*domain.Chapter, config *domain.ConvertConfig, dstPath string) {
	stateManager := state.GetManager()
	stateManager.StartTransaction(config.Id, dstPath, c.getTransactionSize(chapters))

	progressChan := make(chan int64)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		resultPath, err := c.runConversion(ctx, chapters, config, dstPath, progressChan)
		if err != nil {
			c.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Error: %s", err.Error()))
			if canceled := ctx.Err(); canceled != nil {
				c.pushNotifier.Send(config.NotifyToken, "Canceled", fmt.Sprintf("%s conversion canceled", config.Title))
				stateManager.DeleteTransaction(config.Id)
			} else {
				stateManager.SetError(config.Id, err)
			}
			return
		}
		stateManager.SetResultPath(config.Id, resultPath)
		if config.Cloud {
			c.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(resultPath)))
			gCloud, err := cloud.New(config.CloudToken, config.CloudFolder)

			if err != nil {
				c.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(resultPath)))
				return
			}
			gCloud.Upload(resultPath)
		} else {
			c.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s conversion ready", filepath.Base(resultPath)))
		}
	}()

	for processedSize := range progressChan {
		if updated := stateManager.UpdateProgress(config.Id, processedSize); !updated {
			cancel()
			return
		}
	}

	cancel()
	stateManager.SetDone(config.Id)
}

func (c *ConvertMangaUC) runConversion(
	ctx context.Context,
	chapters []*domain.Chapter,
	config *domain.ConvertConfig,
	dstPath string,
	progressChan chan int64,
) (string, error) {
	defer close(progressChan)
	if err := c.setParams(config); err != nil {
		return "", err
	}

	builder := epubBuilder.New()
	builder.SetSettings(c.imageSettings, c.profile)
	builder.Start(config.Title, dstPath)

	for _, chapter := range chapters {
		for pIdx, pagePath := range chapter.PagePaths {
			if err := ctx.Err(); err != nil {
				return "", fmt.Errorf("Job canceled")
			}
			if filepath.Ext(pagePath) == ".xml" {
				continue
			}
			page, err := c.processPage(pagePath, pIdx+1)
			if err != nil {
				return "", err
			}
			builder.AddPage(page, pIdx == 0)
		}
		progressChan <- chapter.Size
	}

	path, err := builder.Build()
	return path, err
}

func (c *ConvertMangaUC) getTransactionSize(chapters []*domain.Chapter) int64 {
	var res int64
	for _, chapter := range chapters {
		res += chapter.Size
	}
	return res
}

func (c *ConvertMangaUC) setParams(config *domain.ConvertConfig) error {
	c.config = config
	profile, err := domain.NewProfile(config.Profile)
	if err != nil {
		return err
	}
	c.profile = profile
	return nil
}

func (c *ConvertMangaUC) processPage(path string, idx int) (*domain.Page, error) {
	page := domain.NewPage(path)
	editor, err := image.NewEditor(
		path,
		c.profile.Width,
		c.profile.Height,
		c.imageSettings.ForceColor,
	)
	if err != nil {
		return nil, err
	}

	page.HasWhiteBg = editor.HasWhiteBg()
	editor.CropMargins()

	isColor := c.imageSettings.ForceColor && editor.IsColored()
	if !isColor {
		editor.Grayscale()
	}
	if c.imageSettings.RemoveRainbowEffect && isColor {
		editor.RemoveRainbowEffect()
	}

	partEditors := editor.TrySplit(c.imageSettings.SpreadSplitter == 2)
	if c.imageSettings.SpreadSplitter != 1 && len(partEditors) > 2 {
		partEditors = partEditors[:2]
	}

	for _, partEditor := range partEditors {
		partEditor.Resize()
		part := domain.NewPagePart(
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
