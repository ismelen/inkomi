package samples

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

var _ ports.ISampleSource = (*CloudSampleSource)(nil)

const downloadTimeout = 30 * time.Second

// CloudSampleSource serves the preview samples stored in cloud storage as id/filename, keeping
// them in memory after the first download. Failed downloads aren't cached.
type CloudSampleSource struct {
	cloudStorage cloud.ICloudStorage
	id           string
	filenames    map[models.PreviewSample]string
	maxSize      int64

	group singleflight.Group
	mu    sync.RWMutex
	cache map[models.PreviewSample][]byte
}

func NewCloudSampleSource(
	cloudStorage cloud.ICloudStorage,
	id string,
	filenames map[models.PreviewSample]string,
	maxSize int64,
) *CloudSampleSource {
	return &CloudSampleSource{
		cloudStorage: cloudStorage,
		id:           id,
		filenames:    filenames,
		maxSize:      maxSize,
		cache:        make(map[models.PreviewSample][]byte),
	}
}

// Get returns the sample's bytes. Callers must not modify them, since they're shared.
func (c *CloudSampleSource) Get(ctx context.Context, sample models.PreviewSample) ([]byte, error) {
	filename, ok := c.filenames[sample]
	if !ok {
		return nil, models.ErrUnknownSample
	}

	c.mu.RLock()
	data, ok := c.cache[sample]
	c.mu.RUnlock()
	if ok {
		return data, nil
	}

	result, err, _ := c.group.Do(string(sample), func() (any, error) {
		// shared by every concurrent caller, so it must not depend on the first one's request
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), downloadTimeout)
		defer cancel()
		data, err := c.download(ctx, filename)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		c.cache[sample] = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

func (c *CloudSampleSource) download(ctx context.Context, filename string) ([]byte, error) {
	reader, err := c.cloudStorage.Download(ctx, c.id, filename)
	if err != nil {
		return nil, fmt.Errorf("download %s/%s: %w", c.id, filename, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(io.LimitReader(reader, c.maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s/%s: %w", c.id, filename, err)
	}
	if int64(len(data)) > c.maxSize {
		return nil, fmt.Errorf("%s/%s: %w", c.id, filename, models.ErrSampleTooLarge)
	}
	return data, nil
}
