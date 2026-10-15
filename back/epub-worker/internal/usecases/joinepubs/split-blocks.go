package joinepubs

import "github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"

// SplitBlocks groups items in order into blocks whose total size never exceeds maxSize.
// An item bigger than maxSize goes alone in its own block.
func SplitBlocks(items []models.JoinItem, maxSize int64) [][]models.JoinItem {
	var blocks [][]models.JoinItem
	var current []models.JoinItem
	var currentSize int64

	for _, item := range items {
		if len(current) > 0 && currentSize+item.Size > maxSize {
			blocks = append(blocks, current)
			current, currentSize = nil, 0
		}
		current = append(current, item)
		currentSize += item.Size
	}

	if len(current) > 0 {
		blocks = append(blocks, current)
	}
	return blocks
}
