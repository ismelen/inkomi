package FileUtils

import (
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	
	return info.IsDir(), nil
}

func IsImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

func HasCorrectExtension(source string, extensions ...string) bool {
	current := strings.ToLower(filepath.Ext(source))
	for _, ext := range extensions {
		if ext == current {
			return true
		}
	}

	return false
}

func FilenameCmp(a, b string) int {
	re := regexp.MustCompile(`(\d+|\D+)`)
	aParts := re.FindAllString(a, -1)
	bParts := re.FindAllString(b, -1)

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, aIsNum := strconv.Atoi(aParts[i])
		bNum, bIsNum := strconv.Atoi(bParts[i])

		// Both are numbers - compare numerically
		if aIsNum == nil && bIsNum == nil {
			if aNum != bNum {
				return aNum - bNum
			}
			continue
		}

		// One is number, one is not - numbers come first
		if aIsNum == nil {
			return 1
		}
		if bIsNum == nil {
			return 1
		}

		// Both are text - compare lexicographically
		if aParts[i] != bParts[i] {
			return strings.Compare(aParts[i], bParts[i])
		}
	}

	// If all parts matched, shorter string comes first
	return len(aParts) - len(bParts)
}

func SaveWithCodec(img *image.Image, targetPath string) (string, error) {
	targetPath += ".jpg"
	err := imaging.Save((*img), targetPath, imaging.JPEGQuality(85))
	if err != nil {
		return "", err
	}
	return targetPath, nil
}
