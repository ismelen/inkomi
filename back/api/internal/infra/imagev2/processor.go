package imagev2

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	// standard library image decoders
	_ "image/jpeg"

	"github.com/h2non/bimg"

	"ismelen/inkomi/internal/domain/manga"
)

// ensure processorImpl implements the domain interface
var _ manga.ImageProcessor = (*processorImpl)(nil)

type processorImpl struct {
	pool *WorkerPool
}

func NewProcessor(workers int) manga.ImageProcessor {
	pool := NewWorkerPool(workers)
	proc := &processorImpl{
		pool: pool,
	}
	pool.Start(proc)
	return proc
}

func (p *processorImpl) ProcessPage(ctx context.Context, imgData []byte, opts manga.ProcessOptions) ([][]byte, error) {
	return p.pool.Submit(ctx, imgData, opts)
}

func (p *processorImpl) processInternal(imgData []byte, opts manga.ProcessOptions) ([][]byte, error) {
	// 1. Decodificación y Split
	bImg := bimg.NewImage(imgData)
	size, err := bImg.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	var buffers [][]byte

	// Evaluate aspect ratio for double page
	aspectRatio := float64(size.Width) / float64(size.Height)
	if aspectRatio >= 1.15 {
		if opts.RotateDoublePage {
			// Rotate 90 degrees
			rotated, err := bImg.Process(bimg.Options{Rotate: 90})
			if err != nil {
				return nil, fmt.Errorf("failed to rotate image: %w", err)
			}
			buffers = append(buffers, rotated)
		} else {
			// Split vertically by half
			halfWidth := size.Width / 2
			left, err := bImg.Extract(0, 0, halfWidth, size.Height)
			if err != nil {
				return nil, fmt.Errorf("failed to extract left half: %w", err)
			}
			right, err := bimg.NewImage(imgData).Extract(halfWidth, 0, halfWidth, size.Height)
			if err != nil {
				return nil, fmt.Errorf("failed to extract right half: %w", err)
			}
			buffers = append(buffers, right, left) // Right page first for manga? Or left first? Let's just process both.
		}
	} else {
		buffers = append(buffers, imgData) // Or bImg.Image()
	}

	var finalBuffers [][]byte

	for _, buf := range buffers {
		processedBuf, err := processSinglePage(buf, opts)
		if err != nil {
			return nil, err
		}
		finalBuffers = append(finalBuffers, processedBuf)
	}

	return finalBuffers, nil
}

func processSinglePage(imgData []byte, opts manga.ProcessOptions) ([]byte, error) {
	// 1. Opciones de Recorte (Crop)
	bImg := bimg.NewImage(imgData)
	
	if opts.Crop.Enabled {
		// Try to extract raw pixels to calculate bounding box
		goImg, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			pngBuf, err2 := bImg.Process(bimg.Options{Type: bimg.PNG})
			if err2 == nil {
				goImg, _, _ = image.Decode(bytes.NewReader(pngBuf))
			}
		}

		if goImg != nil {
			cropRect := calculateCropBox(goImg, opts.Crop.Tolerance)
			
			if cropRect.Dx() > 0 && cropRect.Dy() > 0 {
				imgData, err = bImg.Extract(cropRect.Min.Y, cropRect.Min.X, cropRect.Dx(), cropRect.Dy())
				if err != nil {
					return nil, fmt.Errorf("failed to extract crop: %w", err)
				}
				bImg = bimg.NewImage(imgData)
			}
		}
	}

	// 2. Redimensionado (Scaling)
	targetW := opts.TargetWidth
	if targetW == 0 {
		targetW = 1448
	}
	targetH := opts.TargetHeight
	if targetH == 0 {
		targetH = 1072
	}

	imgData, err := bimg.Resize(imgData, bimg.Options{
		Width:        targetW,
		Height:       targetH,
		Enlarge:      opts.Resize.Enlarge,
		Embed:        opts.Resize.Embed,
		Interpolator: bimg.Lanczos3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resize: %w", err)
	}
	bImg = bimg.NewImage(imgData)

	// 3 & 4. Color Correction & Dithering
	if opts.Color.Enabled || opts.Dither.Enabled {
		pngBuf, err := bImg.Process(bimg.Options{Type: bimg.PNG})
		if err != nil {
			return nil, fmt.Errorf("failed to convert to PNG for processing: %w", err)
		}

		goImg, _, err := image.Decode(bytes.NewReader(pngBuf))
		if err != nil {
			return nil, fmt.Errorf("failed to decode resized png: %w", err)
		}

		bounds := goImg.Bounds()
		rgba := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rgba.Set(x, y, goImg.At(x, y))
			}
		}

		if opts.Color.Enabled {
			applyLevelsAndColor(rgba, opts.Color)
		}

		if opts.Dither.Enabled {
			applyAtkinsonDither(rgba)
		}

		var outBuf bytes.Buffer
		if err := png.Encode(&outBuf, rgba); err != nil {
			return nil, fmt.Errorf("failed to encode processed image: %w", err)
		}
		imgData = outBuf.Bytes()
		bImg = bimg.NewImage(imgData)
	}

	// 5 & 6. Unsharp Mask (Sharpen) & Output
	bimgOpts := bimg.Options{
		Type:    bimg.ImageType(opts.Output.Format),
		Palette: opts.Output.Palette,
		Quality: opts.Output.Quality,
	}
	
	if bimgOpts.Type == 0 {
		bimgOpts.Type = bimg.PNG
	}

	if opts.Sharpen.Enabled {
		bimgOpts.Sharpen = bimg.Sharpen{
			Radius: opts.Sharpen.Radius,
			X1:     opts.Sharpen.X1,
			Y2:     opts.Sharpen.Y2,
			Y3:     opts.Sharpen.Y3,
			M1:     opts.Sharpen.M1,
			M2:     opts.Sharpen.M2,
		}
	}

	finalBuf, err := bImg.Process(bimgOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to process final image: %w", err)
	}

	return finalBuf, nil
}
