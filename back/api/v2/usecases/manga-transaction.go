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
	"log"
	"os"
	"path/filepath"
)

type MangaTransactionUC struct {
	profile       *domain.Profile
	imageSettings *domain.ImageSettings
	pushNotifier  ports.PushNotifier
}

func NewMangaTransactionUC(pushNotifier ports.PushNotifier) *MangaTransactionUC {
	return &MangaTransactionUC{
		imageSettings: domain.NewDefaultImageSettings(),
		pushNotifier:  pushNotifier,
	}
}

func (m *MangaTransactionUC) Execute(chapters []*domain.Chapter, config *domain.TransactionConfig, dstPath string) {
	stateManager := state.GetManager()
	stateManager.StartTransaction(config.Id, dstPath, m.getTransactionSize(chapters))

	progressChan := make(chan int64)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		resultPath, err := m.runConversion(ctx, chapters, config, dstPath, progressChan)
		if err != nil {
			if canceled := ctx.Err(); canceled != nil {
				m.pushNotifier.Send(config.NotifyToken, "Canceled", fmt.Sprintf("%s conversion canceled", config.Title))
				stateManager.DeleteTransaction(config.Id)
			} else {
				m.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Error: %s", err.Error()))
				stateManager.SetError(config.Id, err)
			}
			return
		}
		stateManager.SetResultPath(config.Id, resultPath)

		if config.Cloud {
			m.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("Sending %s to cloud", filepath.Base(resultPath)))
			gCloud, err := cloud.New(config.CloudToken, config.CloudFolder)

			if err != nil {
				m.pushNotifier.Send(config.NotifyToken, "Error", fmt.Sprintf("Cannot send %s to cloud", filepath.Base(resultPath)))
				return
			}
			gCloud.Upload(resultPath)
		} else {
			m.pushNotifier.Send(config.NotifyToken, "Success", fmt.Sprintf("%s transaction ready", filepath.Base(resultPath)))
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

func (m *MangaTransactionUC) runConversion(
	ctx context.Context,
	chapters []*domain.Chapter,
	config *domain.TransactionConfig,
	dstPath string,
	progressChan chan int64,
) (string, error) {
	defer close(progressChan)
	if err := m.setParams(config); err != nil {
		return "", err
	}

	builder := epubBuilder.New()
	builder.SetSettings(m.imageSettings, m.profile)
	builder.Start(config.Title, dstPath)

	for _, chapter := range chapters {
		for pIdx, pagePath := range chapter.GetPagePaths() {
			if err := ctx.Err(); err != nil {
				return "", fmt.Errorf("Job canceled")
			}
			if filepath.Ext(pagePath) == ".xml" {
				continue
			}
			page, err := m.processPage(pagePath, pIdx+1)
			if err != nil {
				return "", err
			}
			builder.AddPage(page, pIdx == 0)
		}
		progressChan <- chapter.Size
	}

	path, err := builder.Build()
	defer os.RemoveAll(filepath.Join(dstPath, "chapters"))

	if err != nil {
		return "", err
	}

	if m.profile.IsKepub {
		kPath, err := ConvertToKepub(path, dstPath, config.Title)
		if err := os.RemoveAll(path); err != nil {
			log.Println(err.Error())
		}
		return kPath, err
	}
	return path, err
}

func (m *MangaTransactionUC) getTransactionSize(chapters []*domain.Chapter) int64 {
	var res int64
	for _, chapter := range chapters {
		res += chapter.Size
	}
	return res
}

func (m *MangaTransactionUC) setParams(config *domain.TransactionConfig) error {
	profile, err := domain.NewProfile(config.Profile)
	if err != nil {
		return err
	}
	m.profile = profile
	return nil
}

func (m *MangaTransactionUC) processPage(path string, idx int) (*domain.Page, error) {
	page := domain.NewPage(path)
	editor, err := image.NewEditor(
		path,
		m.profile.Width,
		m.profile.Height,
		m.imageSettings.ForceColor,
	)
	if err != nil {
		return nil, err
	}

	page.HasWhiteBg = editor.HasWhiteBg()
	editor.CropMargins()

	isColor := m.imageSettings.ForceColor && editor.IsColored()
	if !isColor {
		editor.Grayscale()
	}
	if m.imageSettings.RemoveRainbowEffect && isColor {
		editor.RemoveRainbowEffect()
	}

	partEditors := editor.TrySplit(m.imageSettings.SpreadSplitter == 2)
	if m.imageSettings.SpreadSplitter != 1 && len(partEditors) > 2 {
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
