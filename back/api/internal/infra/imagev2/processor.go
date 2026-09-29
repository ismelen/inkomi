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
	// Step 2: Auto-Crop
	bImg := bimg.NewImage(imgData)
	size, _ := bImg.Size()

	// Decode with stdlib to get raw pixels
	goImg, _, err := image.Decode(bytes.NewReader(imgData))
	if err == nil {
		cropRect := calculateCropBox(goImg)
		// Extract
		if cropRect.Dx() > 0 && cropRect.Dy() > 0 {
			imgData, err = bImg.Extract(cropRect.Min.Y, cropRect.Min.X, cropRect.Dx(), cropRect.Dy())
			if err != nil {
				return nil, fmt.Errorf("failed to extract crop: %w", err)
			}
			bImg = bimg.NewImage(imgData)
		}
	} else {
		// if we can't decode it with stdlib, let bimg convert it to png first
		pngBuf, err := bImg.Process(bimg.Options{Type: bimg.PNG})
		if err == nil {
			goImg, _, err = image.Decode(bytes.NewReader(pngBuf))
			if err == nil {
				cropRect := calculateCropBox(goImg)
				if cropRect.Dx() > 0 && cropRect.Dy() > 0 {
					imgData, err = bImg.Extract(cropRect.Min.Y, cropRect.Min.X, cropRect.Dx(), cropRect.Dy())
					if err != nil {
						return nil, fmt.Errorf("failed to extract crop: %w", err)
					}
					bImg = bimg.NewImage(imgData)
				}
			}
		}
	}

	// Step 3: Scaling (Resize with bimg, fit within target dims, Lanczos3)
	targetW := opts.TargetWidth
	if targetW == 0 {
		targetW = 1448
	}
	targetH := opts.TargetHeight
	if targetH == 0 {
		targetH = 1072
	}

	imgData, err = bimg.Resize(imgData, bimg.Options{
		Width:        targetW,
		Height:       targetH,
		Enlarge:      false,
		Embed:        false,
		Interpolator: bimg.Lanczos3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resize: %w", err)
	}
	bImg = bimg.NewImage(imgData)

	// We need PNG for manual processing to avoid decoding issues
	pngBuf, err := bImg.Process(bimg.Options{Type: bimg.PNG})
	if err != nil {
		return nil, fmt.Errorf("failed to convert to PNG for processing: %w", err)
	}

	goImg, _, err = image.Decode(bytes.NewReader(pngBuf))
	if err != nil {
		return nil, fmt.Errorf("failed to decode resized png: %w", err)
	}

	// Step 4 & 6: Levels & Color Correction & Dithering
	// Extract raw pixels to a modifiable format (RGBA)
	bounds := goImg.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, goImg.At(x, y))
		}
	}

	// Apply levels and dithering on raw pixels
	applyLevelsAndColor(rgba)
	applyAtkinsonDither(rgba)

	// Encode back to PNG
	var outBuf bytes.Buffer
	if err := png.Encode(&outBuf, rgba); err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	// Step 5: Unsharp Mask (bimg)
	bImgProcessed := bimg.NewImage(outBuf.Bytes())

	// Unsharp mask options. We can use bimg's Sharpen.
	options := bimg.Options{
		Sharpen: bimg.Sharpen{Radius: 1, X1: 1.5, Y2: 20, Y3: 50, M1: 0, M2: 3}, // Custom unsharp mask
		Type:    bimg.PNG,
		Palette: true, // 8-bit indexed PNG (quantizes up to 256 colors)
	}

	// Step 7: Codificación PNG Indexado
	finalBuf, err := bImgProcessed.Process(options)
	if err != nil {
		return nil, fmt.Errorf("failed to process final image: %w", err)
	}

	return finalBuf, nil
}
