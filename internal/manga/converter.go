package manga

import (
	documentBuilder "ismelen/ermc/internal/document-builder"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/pkg"
	"log"
	"os"
	"runtime"

	"golang.org/x/sync/errgroup"
)

type converter struct {
	settings        *domain.Settings
	pageProcessor   *PageProcessor
	documentBuilder documentBuilder.BuilderI
	ramLimit        int64 // MB
	resultChan chan string
}


func NewConverter(settings *domain.Settings, ramLimit int64, resultChan chan string) *converter {
	return &converter{
		settings:        settings,
		ramLimit:        ramLimit,
		pageProcessor: NewPageProcessor(settings),
		resultChan: resultChan,
	}
}

func (c *converter) Convert(format string) ([]string, error) {
	defer os.RemoveAll(c.settings.Output.Chapters)
	
	jobChan := make(chan func())
	c.launchPageWorkers(jobChan)

	var buildGroup errgroup.Group
	results := pkg.NewSyncList()

	for i, volume := range c.settings.Volumes {
		vol := volume
		vol.Wg.Add(1)

		go func(v *domain.Volume) {
			defer v.Wg.Done()
			for _, chapter := range v.Chapters {
				v.Wg.Add(len(chapter.Pages))
				for i, page := range chapter.Pages {
					jobChan <- func(){
						err := c.pageProcessor.ProcessNewPage(page, i+1)
						if err != nil {
							log.Fatal(err)
						}
						v.Wg.Done()
					}
				}
			}
		}(vol)

		idx := i
		buildGroup.Go(func() error {
			vol.Wg.Wait()
			path, err := c.getOutput(vol, format)
			c.settings.Volumes[idx] = nil
			runtime.GC()

			if err != nil {
				return err
			}

			c.resultChan <- path
			results.Add(path)
			return nil
		})
	}

	if err := buildGroup.Wait(); err != nil {
		return nil, err
	}
	close(jobChan)
	close(c.resultChan)
	
	return results.Values, nil
}

func (c *converter) getOutput(volume *domain.Volume, format string) (string, error) {
	db, err := documentBuilder.GetBuilder(format)
	if err != nil {
		return "", err
	}

	db.SetSettings(c.settings)
	db.Start(volume.Name)
	
	for _, chapter := range volume.Chapters {
		for i, page := range chapter.Pages {
			db.AddPage(page, i == 0)
			page.Parts = nil
		}
	}

	return db.Build()
}

func (c *converter) launchPageWorkers(jobChan chan func()) {
	threadAmount := int(c.ramLimit / 60)
	if threadAmount == 0 {
		threadAmount = 1
	}

	for range threadAmount {
		go func() {
			for job := range jobChan {
				job()
			}
		}()
	}
}
