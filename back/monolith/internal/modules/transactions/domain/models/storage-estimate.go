package models

import "math"

// StorageItem is a file of an upload as the storage estimate sees it.
type StorageItem struct {
	// Size is the declared size; 0 for library items, whose size is unknown.
	Size      int64
	IsLibrary bool
	// IsEpub marks files uploaded as epub, which aren't converted.
	IsEpub bool
}

// StorageRequest is one root of an upload: a loose file or a folder with its items.
type StorageRequest struct {
	Items []StorageItem
	Join  bool
	Kepub bool
}

type StorageEstimateOptions struct {
	// EpubRatio is the size of the epub of a manga file relative to the file.
	EpubRatio float64
	// LibraryEstimate is the size assumed for a library item.
	LibraryEstimate int64
	// Concurrency is how many files are converted at once.
	Concurrency int
	// MaxJoinBlock is the largest block a join writes.
	MaxJoinBlock int64
}

// EstimateStoragePeak returns the most bytes the root needs on disk at once. Every step deletes
// its input as soon as its output is stored (source → epub → joined block → kepub), so the
// estimate is the largest of its phases instead of the sum of every file:
//   - conversion: each file is either its source or its epub, and the ones being converted hold both;
//   - join: every epub plus the block being written (each block's chapters are deleted once it's stored);
//   - kepub: every epub or block plus the output of the one being converted.
func EstimateStoragePeak(req StorageRequest, opts StorageEstimateOptions) int64 {
	if len(req.Items) == 0 {
		return 0
	}
	ratio := math.Max(opts.EpubRatio, 1)

	var settled, epubs, largestEpub int64
	for _, item := range req.Items {
		size := item.Size
		if item.IsLibrary || size <= 0 {
			size = opts.LibraryEstimate
		}
		epub := size
		if !item.IsEpub && !item.IsLibrary {
			epub = int64(math.Ceil(float64(size) * ratio))
		}
		settled += max(size, epub)
		epubs += epub
		largestEpub = max(largestEpub, epub)
	}

	inFlight := int64(min(max(1, opts.Concurrency), len(req.Items)))
	peak := settled + inFlight*largestEpub

	largestOutput := largestEpub
	if req.Join {
		block := epubs
		if opts.MaxJoinBlock > 0 {
			block = min(block, opts.MaxJoinBlock)
		}
		peak = max(peak, epubs+block)
		largestOutput = block
	}
	if req.Kepub {
		peak = max(peak, epubs+largestOutput)
	}
	return peak
}
