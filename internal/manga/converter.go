package manga

import (
	documentBuilder "ismelen/ermc/internal/document-builder"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/pkg"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

type converter struct {
	settings        *domain.Settings
	pageProcessor   *PageProcessor
	documentBuilder documentBuilder.BuilderI
	ramLimit        int64 // MB
}

type pageTask struct {
	page *domain.Page
	num  int
	wg   *sync.WaitGroup
	db   documentBuilder.BuilderI
}

func NewConverter(settings *domain.Settings, documentBuilder documentBuilder.BuilderI, ramLimit int64) *converter {
	return &converter{
		settings:        settings,
		ramLimit:        ramLimit,
		documentBuilder: documentBuilder,
	}
}

func (c *converter) Convert() ([]string, error) {
	pageChan := make(chan pageTask)
	c.launchPageWorkers(pageChan)
	var buildGroup errgroup.Group
	results := pkg.NewSyncList()

	volDocumentBuilders := map[string]documentBuilder.BuilderI{}

	for i, vol := range c.settings.Volumes {
		vol.Wg.Add(1)
		volDocumentBuilders[vol.Name] = c.documentBuilder.Copy().Start(vol)

		go func(v *domain.Volume) {
			defer v.Wg.Done()
			for _, chapter := range v.Chapters {
				for i, page := range chapter.Pages {
					v.Wg.Add(1)
					pageChan <- pageTask{
						page: page,
						num:  i + 1,
						wg:   v.Wg,
						db:   volDocumentBuilders[v.Name],
					}
				}
			}
		}(vol)

		idx := i
		buildGroup.Go(func() error {
			vol.Wg.Wait()
			path, err := volDocumentBuilders[vol.Name].Build()

			delete(volDocumentBuilders, vol.Name)
			c.settings.Volumes[idx] = nil
			runtime.GC()

			if err != nil {
				return err
			}

			results.Add(path)
			return nil
		})
	}

	if err := buildGroup.Wait(); err != nil {
		return nil, err
	}
	close(pageChan)

	return results.Values, nil
}

func (c *converter) launchPageWorkers(pageChan chan pageTask) {
	//TODO: Change to read RAM available
	threadAmount := int(c.ramLimit / 200)
	if threadAmount == 0 {
		threadAmount = 100
	}

	for i := range threadAmount {
		go func(id int) {
			for job := range pageChan {
				c.pageProcessor.ProcessNewPage(job.page, job.num)
				runtime.GC()
				job.db.AddPage(job.page)
				job.wg.Done()
			}
		}(i)
	}
}
