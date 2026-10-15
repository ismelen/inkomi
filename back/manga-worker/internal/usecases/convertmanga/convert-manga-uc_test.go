package convertmanga_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/convertmanga"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type published struct {
	subject string
	msg     jobs.Msg
}

type fixture struct {
	queue     *mocks.MockQueue
	cloud     *mocks.MockCloudStorage
	extractor *mocks.MockPageExtractor
	processor *mocks.MockPageProcessor
	builder   *mocks.MockMangaEpubBuilder

	mu        sync.Mutex
	events    []published
	uploaded  []string
	deleted   []string
	built     []models.EpubPage
	contents  []string
	builtMeta models.EpubMeta
}

func newFixture(t *testing.T, pages ...string) *fixture {
	t.Helper()

	f := &fixture{}
	f.queue = &mocks.MockQueue{PublishFn: func(_ context.Context, subject string, payload any) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.events = append(f.events, published{subject: subject, msg: payload.(jobs.Msg)})
		return nil
	}}
	f.cloud = &mocks.MockCloudStorage{
		DownloadFileFn: func(_ context.Context, _, _, path string) error {
			return os.WriteFile(path, []byte("archive"), 0o644)
		},
		UploadFileFn: func(_ context.Context, id, filename, _ string) (int64, error) {
			f.uploaded = append(f.uploaded, id+"/"+filename)
			return 42, nil
		},
		DeleteFn: func(_ context.Context, id, filename string) error {
			f.deleted = append(f.deleted, id+"/"+filename)
			return nil
		},
	}
	f.extractor = &mocks.MockPageExtractor{ExtractFn: func(_ context.Context, _, dstDir string) ([]string, error) {
		paths := make([]string, len(pages))
		for i, content := range pages {
			paths[i] = filepath.Join(dstDir, fmt.Sprintf("%04d.jpg", i+1))
			if err := os.WriteFile(paths[i], []byte(content), 0o644); err != nil {
				return nil, err
			}
		}
		return paths, nil
	}}
	f.processor = &mocks.MockPageProcessor{}
	f.builder = &mocks.MockMangaEpubBuilder{BuildFn: func(outPath string, meta models.EpubMeta, pages []models.EpubPage) error {
		f.builtMeta, f.built = meta, pages
		for _, page := range pages {
			data, err := os.ReadFile(page.Path)
			if err != nil {
				return err
			}
			f.contents = append(f.contents, string(data))
		}
		return os.WriteFile(outPath, nil, 0o644)
	}}
	return f
}

func (f *fixture) uc() *convertmanga.ConvertMangaUC {
	extractors := map[models.SourceFormat]ports.IPageExtractor{
		models.SourceFormatCBZ: f.extractor,
		models.SourceFormatPDF: f.extractor,
	}
	return convertmanga.NewConvertMangaUC(f.queue, f.cloud, extractors, f.processor, f.builder, 4)
}

func (f *fixture) subjects() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	subjects := make([]string, len(f.events))
	for i, e := range f.events {
		subjects[i] = e.subject
	}
	return subjects
}

func mangaMsg() jobs.Msg {
	return jobs.Msg{
		Id:       "src-1",
		UserId:   7,
		Filename: "chapter 1.cbz",
		Title:    "Chapter 1",
		Manga:    &jobs.Manga{Rtl: true, MangaCfg: json.RawMessage(`{"targetWidth":600,"targetHeight":800}`)},
	}
}

func TestConvertMangaUC_Execute_Valid_PublishesNextSubject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		modify func(*jobs.Msg)
		want   jobs.Subject
	}{
		{"done", func(*jobs.Msg) {}, jobs.SubjectDone},
		{"kepub", func(m *jobs.Msg) { m.Kepubify = true }, jobs.SubjectKepub},
		{"join", func(m *jobs.Msg) {
			m.Kepubify = true
			m.Parent = &jobs.Parent{Id: "folder", Total: 2}
		}, jobs.SubjectJoin},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(t, "p1", "p2")
			msg := mangaMsg()
			c.modify(&msg)

			// Act
			err := f.uc().Execute(context.Background(), msg, jobrun.Delivery{})

			// Assert
			require.NoError(t, err)
			assert.Equal(t, []string{string(jobs.SubjectStart), string(c.want)}, f.subjects())
			result := f.events[1].msg
			assert.Equal(t, "Chapter 1.epub", result.Filename)
			assert.Equal(t, int64(42), result.Size)
			assert.Equal(t, "src-1", result.Id)
			assert.Nil(t, result.Manga)
			assert.Equal(t, []string{"src-1/Chapter 1.epub"}, f.uploaded)
			assert.Equal(t, []string{"src-1/chapter 1.cbz"}, f.deleted)
			assert.Equal(t, models.EpubMeta{Title: "Chapter 1", RTL: true, Width: 600, Height: 800}, f.builtMeta)
		})
	}
}

func TestConvertMangaUC_Execute_SplitPages_KeepsReadingOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(t, "p1", "spread", "p3")
	f.processor.ProcessFn = func(_ context.Context, data []byte, _ models.MangaConfig, rtl bool) ([]models.ProcessedPage, error) {
		if string(data) != "spread" {
			return []models.ProcessedPage{{Data: data, Format: models.ImageFormatJPEG}}, nil
		}
		assert.True(t, rtl)
		return []models.ProcessedPage{
			{Data: []byte("spread-right"), Split: models.SplitToRight, Format: models.ImageFormatPNG},
			{Data: []byte("spread-left"), Split: models.SplitToLeft, Format: models.ImageFormatPNG},
		}, nil
	}

	// Act
	err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"p1", "spread-right", "spread-left", "p3"}, f.contents)
	assert.Equal(t, models.SplitToRight, f.built[1].Split)
	assert.Equal(t, models.ImageFormatPNG, f.built[1].Format)
}

func TestConvertMangaUC_Execute_NoMangaInfo_UsesDefaultsAndLTR(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(t, "p1")
	msg := mangaMsg()
	msg.Manga = nil
	msg.Title = ""

	// Act
	err := f.uc().Execute(context.Background(), msg, jobrun.Delivery{})

	// Assert
	require.NoError(t, err)
	def := models.DefaultMangaConfig()
	assert.Equal(t, models.EpubMeta{Title: "chapter 1", RTL: false, Width: def.TargetWidth, Height: def.TargetHeight}, f.builtMeta)
}

func TestConvertMangaUC_Execute_InvalidInput_ReturnsPermanentWithoutWork(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		modify func(*jobs.Msg)
	}{
		{"invalid config", func(m *jobs.Msg) { m.MangaCfg = json.RawMessage(`{"gamma":9}`) }},
		{"unsupported format", func(m *jobs.Msg) { m.Filename = "book.epub" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(t, "p1")
			downloaded := false
			f.cloud.DownloadFileFn = func(context.Context, string, string, string) error {
				downloaded = true
				return nil
			}
			msg := mangaMsg()
			c.modify(&msg)

			// Act
			err := f.uc().Execute(context.Background(), msg, jobrun.Delivery{})

			// Assert
			require.Error(t, err)
			assert.True(t, jobrun.IsPermanent(err))
			assert.False(t, downloaded)
			assert.Empty(t, f.subjects())
		})
	}
}

func TestConvertMangaUC_Execute_DownloadFails_DependsOnExistingEpub(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		epubExists bool
		wantErr    bool
	}{
		{"already converted", true, false},
		{"not converted", false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(t, "p1")
			f.cloud.DownloadFileFn = func(context.Context, string, string, string) error {
				return errors.New("not found")
			}
			f.cloud.SizeFn = func(_ context.Context, _, filename string) (int64, error) {
				if c.epubExists && filename == "Chapter 1.epub" {
					return 10, nil
				}
				return 0, errors.New("not found")
			}

			// Act
			err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

			// Assert
			if c.wantErr {
				require.Error(t, err)
				assert.False(t, jobrun.IsPermanent(err))
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, []string{string(jobs.SubjectStart)}, f.subjects())
			assert.Empty(t, f.uploaded)
		})
	}
}

func TestConvertMangaUC_Execute_ConversionFails_ReturnsPermanent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*fixture)
	}{
		{"no pages", func(f *fixture) {
			f.extractor.ExtractFn = func(context.Context, string, string) ([]string, error) { return nil, nil }
		}},
		{"corrupt archive", func(f *fixture) {
			f.extractor.ExtractFn = func(context.Context, string, string) ([]string, error) {
				return nil, errors.New("zip: not a valid zip file")
			}
		}},
		{"corrupt page", func(f *fixture) {
			f.processor.ProcessFn = func(context.Context, []byte, models.MangaConfig, bool) ([]models.ProcessedPage, error) {
				return nil, errors.New("unsupported image format")
			}
		}},
		{"build fails", func(f *fixture) {
			f.builder.BuildFn = func(string, models.EpubMeta, []models.EpubPage) error { return errors.New("disk full") }
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(t, "p1", "p2")
			c.setup(f)

			// Act
			err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

			// Assert
			require.Error(t, err)
			assert.True(t, jobrun.IsPermanent(err))
			assert.Empty(t, f.uploaded)
			assert.Empty(t, f.deleted)
		})
	}
}

func TestConvertMangaUC_Execute_ProcessingCanceled_ReturnsTransientError(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(t, "p1")
	f.processor.ProcessFn = func(context.Context, []byte, models.MangaConfig, bool) ([]models.ProcessedPage, error) {
		return nil, context.Canceled
	}

	// Act
	err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

	// Assert
	require.Error(t, err)
	assert.False(t, jobrun.IsPermanent(err))
}

func TestConvertMangaUC_Execute_UploadFails_ReturnsTransientError(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(t, "p1")
	f.cloud.UploadFileFn = func(context.Context, string, string, string) (int64, error) {
		return 0, errors.New("timeout")
	}

	// Act
	err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

	// Assert
	require.Error(t, err)
	assert.False(t, jobrun.IsPermanent(err))
	assert.Equal(t, []string{string(jobs.SubjectStart)}, f.subjects())
	assert.Empty(t, f.deleted)
}

func TestConvertMangaUC_Execute_DeleteFails_StillSucceeds(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(t, "p1")
	f.cloud.DeleteFn = func(context.Context, string, string) error { return errors.New("timeout") }

	// Act
	err := f.uc().Execute(context.Background(), mangaMsg(), jobrun.Delivery{})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{string(jobs.SubjectStart), string(jobs.SubjectDone)}, f.subjects())
}
