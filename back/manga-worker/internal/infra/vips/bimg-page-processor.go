package vips

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	// decoders for the pages handed to the pure Go steps
	_ "image/jpeg"

	"github.com/h2non/bimg"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/imaging"
)

var _ ports.IPageProcessor = (*BimgPageProcessor)(nil)

// spreadRatio is the width/height ratio from which a page is treated as a double page spread.
const spreadRatio = 1.15

type BimgPageProcessor struct {
	pool *WorkerPool
}

// NewBimgPageProcessor starts workers goroutines that live until ctx is done.
func NewBimgPageProcessor(ctx context.Context, workers int) *BimgPageProcessor {
	// pages are processed once, so libvips' operation cache only wastes memory
	bimg.VipsCacheSetMax(0)
	return &BimgPageProcessor{pool: NewWorkerPool(ctx, workers, processPage)}
}

func (b *BimgPageProcessor) Process(ctx context.Context, data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error) {
	return b.pool.Submit(ctx, data, cfg, rtl)
}

type pagePart struct {
	data  []byte
	split models.SplitOperation
}

func processPage(data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error) {
	parts, err := splitSpread(data, cfg, rtl)
	if err != nil {
		return nil, err
	}

	pages := make([]models.ProcessedPage, 0, len(parts))
	for _, part := range parts {
		page, err := processPart(part, cfg)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}

// splitSpread rotates a double page spread or splits it into its halves in reading order.
func splitSpread(data []byte, cfg models.MangaConfig, rtl bool) ([]pagePart, error) {
	img := bimg.NewImage(data)
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("read image size: %w", err)
	}
	if size.Height == 0 || float64(size.Width)/float64(size.Height) < spreadRatio {
		return []pagePart{{data: data, split: models.SplitNone}}, nil
	}

	if cfg.RotateDoublePage {
		rotated, err := img.Process(bimg.Options{Rotate: bimg.D90})
		if err != nil {
			return nil, fmt.Errorf("rotate spread: %w", err)
		}
		return []pagePart{{data: rotated, split: models.SplitRotated}}, nil
	}

	half := size.Width / 2
	left, err := bimg.NewImage(data).Extract(0, 0, half, size.Height)
	if err != nil {
		return nil, fmt.Errorf("extract left half: %w", err)
	}
	right, err := bimg.NewImage(data).Extract(0, half, size.Width-half, size.Height)
	if err != nil {
		return nil, fmt.Errorf("extract right half: %w", err)
	}

	leftPart := pagePart{data: left, split: models.SplitToLeft}
	rightPart := pagePart{data: right, split: models.SplitToRight}
	if rtl {
		return []pagePart{rightPart, leftPart}, nil
	}
	return []pagePart{leftPart, rightPart}, nil
}

func processPart(part pagePart, cfg models.MangaConfig) (models.ProcessedPage, error) {
	data := part.data

	if cfg.CropEnabled() {
		cropped, err := cropMargins(data, cfg.CropTolerance)
		if err != nil {
			return models.ProcessedPage{}, err
		}
		data = cropped
	}

	data, err := bimg.Resize(data, bimg.Options{
		Width:        cfg.TargetWidth,
		Height:       cfg.TargetHeight,
		Enlarge:      cfg.Enlarge,
		Embed:        cfg.Embed,
		Extend:       bimg.ExtendWhite,
		Interpolator: bimg.Lanczos3,
	})
	if err != nil {
		return models.ProcessedPage{}, fmt.Errorf("resize: %w", err)
	}

	data, err = adjustLevels(data, cfg)
	if err != nil {
		return models.ProcessedPage{}, err
	}

	format := models.ImageFormatJPEG
	opts := bimg.Options{Type: bimg.JPEG, Quality: cfg.Quality}
	if cfg.Palette8bit {
		format = models.ImageFormatPNG
		opts = bimg.Options{Type: bimg.PNG, Palette: true, Quality: cfg.Quality}
	}
	if cfg.Grayscale {
		opts.Interpretation = bimg.InterpretationBW
	}
	if cfg.Sharpen.Enabled {
		opts.Sharpen = bimg.Sharpen{
			Radius: cfg.Sharpen.Radius,
			X1:     cfg.Sharpen.X1,
			Y2:     cfg.Sharpen.Y2,
			Y3:     cfg.Sharpen.Y3,
			M1:     cfg.Sharpen.M1,
			M2:     cfg.Sharpen.M2,
		}
	}

	out, err := bimg.NewImage(data).Process(opts)
	if err != nil {
		return models.ProcessedPage{}, fmt.Errorf("encode page: %w", err)
	}
	size, err := bimg.NewImage(out).Size()
	if err != nil {
		return models.ProcessedPage{}, fmt.Errorf("read page size: %w", err)
	}

	return models.ProcessedPage{
		Data:   out,
		Width:  size.Width,
		Height: size.Height,
		Split:  part.split,
		Format: format,
	}, nil
}

func cropMargins(data []byte, tolerance float64) ([]byte, error) {
	img, err := decode(data)
	if err != nil {
		return nil, err
	}

	box := imaging.CropBox(img, tolerance)
	if box == img.Bounds() {
		return data, nil
	}
	cropped, err := bimg.NewImage(data).Extract(box.Min.Y, box.Min.X, box.Dx(), box.Dy())
	if err != nil {
		return nil, fmt.Errorf("crop margins: %w", err)
	}
	return cropped, nil
}

// adjustLevels runs the pure Go colour steps (levels and dithering), which libvips lacks.
func adjustLevels(data []byte, cfg models.MangaConfig) ([]byte, error) {
	img, err := decode(data)
	if err != nil {
		return nil, err
	}

	rgba := imaging.ToRGBA(img)
	imaging.ApplyLevels(rgba, imaging.Levels{
		Gamma:      cfg.Gamma,
		BlackPoint: cfg.BlackPointThreshold,
		WhitePoint: cfg.WhitePointThreshold,
		Brightness: cfg.Brightness,
		Contrast:   cfg.Contrast,
		Grayscale:  cfg.Grayscale,
	})
	if cfg.Dithering {
		imaging.AtkinsonDither(rgba)
	}

	var buf bytes.Buffer
	// libvips re-encodes it right after, so compressing here is wasted time
	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(&buf, rgba); err != nil {
		return nil, fmt.Errorf("encode adjusted page: %w", err)
	}
	return buf.Bytes(), nil
}

// decode reads data with the Go decoders, converting it to PNG with libvips first when Go
// doesn't support its format (webp, tiff, jp2...).
func decode(data []byte) (image.Image, error) {
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return img, nil
	}

	converted, err := bimg.NewImage(data).Convert(bimg.PNG)
	if err != nil {
		return nil, fmt.Errorf("convert page to png: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(converted))
	if err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	return img, nil
}
