package manga

import (
	"ismelen/ermc/v2/pkg"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

type converter struct {
	settings      *Settings
	pageProcessor *PageProcessor
}

type pageTask struct {
	page *Page
	num  int
	wg   *sync.WaitGroup
}

func New(settings *Settings) *converter {
	return &converter{
		settings: settings,
	}
}

func (c *converter) Convert() ([]string, error) {
	pageChan := make(chan pageTask)
	c.launchPageWorkers(pageChan)
	var buildGroup errgroup.Group
	results := pkg.NewSyncList()

	for _, vol := range c.settings.Volumes {
		vol.Wg.Add(1)
		go func(v *Volume) {
			defer v.Wg.Done()
			for _, chapter := range v.Chapters {
				for i, page := range chapter.Pages {
					v.Wg.Add(1)
					pageChan <- pageTask{
						page: page,
						num:  i + 1,
						wg:   v.Wg,
					}
				}
			}
		}(vol)

		buildGroup.Go(func() error {
			vol.Wg.Wait()
			path, err := c.settings.DocumentProcessor.GetOutput()

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
	threadAmount := runtime.NumCPU() - 2
	if c.settings.LowRAM || threadAmount <= 0 {
		threadAmount = 1
	}

	for i := range threadAmount {
		go func(id int) {
			for job := range pageChan {
				//TODO: implement
				c.pageProcessor.ProcessNewPage(job.page, job.num)

				job.wg.Done()
			}
		}(i)
	}
}
