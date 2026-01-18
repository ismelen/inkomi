package MangaConverter

import (
	"fmt"
	EpubBuilder "ismelen/ermc/internal/manga/logic/builders/epub"
	PageConverter "ismelen/ermc/internal/manga/logic/coverters/page"
	manga "ismelen/ermc/internal/manga/logic/models"
	SharedInterfaces "ismelen/ermc/internal/shared/logic/interfaces"
	SharedModels "ismelen/ermc/internal/shared/logic/models"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

type MangaConverter struct {
	options  *manga.ConverterOptions
	notifier *SharedInterfaces.Notifier
}

func New(options *manga.ConverterOptions, notifier *SharedInterfaces.Notifier) MangaConverter {
	return MangaConverter{
		options:  options,
		notifier: notifier,
	}
}

func (this *MangaConverter) Convert() ([]string, error) {
	vols, err := this.options.GetVolumes()
	if err != nil {
		this.notifier.Notify(MangaConverterEvent{
			Type: EventError,
			Err:  err,
		})
		return nil, err
	}

	pageTasks := make(chan pageTask)
	chaptersDir := filepath.Join(this.options.Output, "chapters")
	var buildGroup errgroup.Group
	results := &SharedModels.SyncList{}

	this.launchPageWorkers(pageTasks)

	for _, vol := range vols {
		vol.Wg.Add(1)

		go func(v manga.Volume) {
			defer vol.Wg.Done()
			for _, chap := range v.Chapters {
				pages, err := chap.GetPages(chaptersDir)
				this.notifier.Notify(MangaConverterEvent{
					Type: EventChapterStart,
					Cant: len(pages),
				})
				if err != nil {
					this.notifier.Notify(MangaConverterEvent{
						Type: EventError,
						Err:  err,
					})
					return
				}

				for i, page := range pages {
					v.Wg.Add(1)
					pageTasks <- pageTask{
						page: page,
						num:  i + 1,
						wg:   v.Wg,
					}
				}
			}
		}(vol)

		buildGroup.Go(func() error {
			vol.Wg.Wait()
			path, err := this.generateOutput(vol.Filename, vol.Chapters...)
			runtime.GC()
			if err != nil {
				return err
			}

			results.Add(path)
			return nil
		})
	}

	if err := buildGroup.Wait(); err != nil {
		this.notifier.Notify(MangaConverterEvent{
			Type: EventError,
			Err:  err,
		})
		return nil, err
	}
	close(pageTasks)

	this.notifier.Notify(MangaConverterEvent{Type: EventDone})
	return results.Values, nil
}

func (this *MangaConverter) generateOutput(dstFileName string, chapters ...*manga.Chapter) (path string, err error) {
	switch this.options.Format {
	case "Auto", "CBZ", "PDF", "EPUB":
		builder := EpubBuilder.New(this.options, dstFileName, chapters...)
		path, err = builder.Build()
	}

	if path == "" {
		return "", fmt.Errorf("Cannot generate output for %s", dstFileName)
	}

	return
}

type pageTask struct {
	page *manga.Page
	num  int
	wg   *sync.WaitGroup
}

func (this *MangaConverter) launchPageWorkers(pageTasks chan pageTask) {
	cant := runtime.NumCPU() - 2
	if this.options.LowRAM || cant <= 0 {
		cant = 1
	}

	for i := range cant {
		go func(id int) {
			for task := range pageTasks {
				fmt.Printf("[Worker %d]\t Working on %s\n", id, task.page.Path)
				converter, err := PageConverter.New(
					task.page,
					this.options,
				)
				if err != nil {
					this.notifier.Notify(MangaConverterEvent{
						Type: EventError,
						Err:  err,
					})
					close(pageTasks)
					continue
				}

				err = converter.Convert(task.num)
				if err != nil {
					this.notifier.Notify(MangaConverterEvent{
						Type: EventError,
						Err:  err,
					})
					close(pageTasks)
					continue
				}

				this.notifier.Notify(MangaConverterEvent{
					Type: EventPageFinished,
				})

				task.wg.Done()
			}
		}(i)
	}
}
