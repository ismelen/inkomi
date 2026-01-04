package cover

// package manga

// import (
// 	"fmt"
// 	"image"
// 	"os"
// 	"path/filepath"

// 	"github.com/disintegration/imaging"
// 	_ "golang.org/x/image/webp"
// )

// type Cover struct {
// 	opt    Options
// 	source string
// 	image  image.Image
// }

// func NewCover(source string, opt Options) (*Cover, error) {
// 	img, err := imaging.Open(source)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open cover image: %w", err)
// 	}

// 	c := &Cover{
// 		opt:    opt,
// 		source: source,
// 		image:  img,
// 	}
// 	c.process()
// 	return c, nil
// }

// func (c *Cover) process() {
// 	// 1. Convert to simple RGB (imaging handles this usually on load or ops)
// 	// 2. AutoContrast
// 	// Python: ImageOps.autocontrast(preserve_tone=True)
// 	// We skip exact implementation or use simple contrast.

// 	if !c.opt.ColorMode {
// 		c.image = imaging.Grayscale(c.image)
// 	}

// 	c.cropMainCover()

// 	targetW, targetH := c.opt.ProfileData.Width, c.opt.ProfileData.Height
// 	// KindleScribeAZW3 logic removed

// 	// Thumbnail logic (resize)
// 	// Python: self.image.thumbnail(tuple(size), Image.Resampling.LANCZOS)
// 	// Thumbnail preserves aspect ratio.
// 	c.image = imaging.Fit(c.image, targetW, targetH, imaging.Lanczos)
// }

// func (c *Cover) cropMainCover() {
// 	w, h := c.image.Bounds().Dx(), c.image.Bounds().Dy()
// 	ratio := float64(w) / float64(h)

// 	var rect image.Rectangle

// 	if ratio > 2 {
// 		if c.opt.Manga {
// 			// crop((w/6, 0, w/2 - w * 0.02, h))
// 			x0 := int(float64(w) / 6.0)
// 			x1 := int(float64(w)/2.0 - float64(w)*0.02)
// 			rect = image.Rect(x0, 0, x1, h)
// 		} else {
// 			// crop((w/2 + w * 0.02, 0, 5/6 * w, h))
// 			x0 := int(float64(w)/2.0 + float64(w)*0.02)
// 			x1 := int(5.0 / 6.0 * float64(w))
// 			rect = image.Rect(x0, 0, x1, h)
// 		}
// 		c.image = imaging.Crop(c.image, rect)

// 	} else if ratio > 1.34 {
// 		if c.opt.Manga {
// 			// crop((0, 0, w/2 - w * 0.03, h))
// 			x1 := int(float64(w)/2.0 - float64(w)*0.03)
// 			rect = image.Rect(0, 0, x1, h)
// 		} else {
// 			// crop((w/2 + w * 0.03, 0, w, h))
// 			x0 := int(float64(w)/2.0 + float64(w)*0.03)
// 			rect = image.Rect(x0, 0, w, h)
// 		}
// 		c.image = imaging.Crop(c.image, rect)
// 	}
// }

// func (c *Cover) SaveToEPUB(target string, tomeID int, lenTomes int) error {
// 	var err error
// 	if tomeID == 0 {
// 		err = imaging.Save(c.image, target, imaging.JPEGQuality(85))
// 	} else {
// 		// Draw text on copy
// 		// Go doesn't have built-in font drawing easily without importing freetype/etc.
// 		// For now, let's just save the image without text or skip text.
// 		// Implementing text drawing requires loading a font file.
// 		// We will skip text for this iteration or add a TODO.
// 		// "draw.text(..., text=f'{tomeid}/{len_tomes}', ...)"

// 		// Just save copy for now.
// 		err = imaging.Save(c.image, target, imaging.JPEGQuality(85))
// 	}
// 	if err != nil {
// 		return fmt.Errorf("failed to save cover to epub: %w", err)
// 	}
// 	return nil
// }

// func (c *Cover) SaveToKindle(kindlePath string, asin string) error {
// 	// Python: self.image = ImageOps.contain(self.image, (300, 470), Image.Resampling.LANCZOS)
// 	c.image = imaging.Fit(c.image, 300, 470, imaging.Lanczos)

// 	// Path construction
// 	// os.path.join(kindle.path.split('documents')[0], 'system', 'thumbnails', 'thumbnail_' + asin + '_EBOK_portrait.jpg')
// 	// We assume kindlePath is the path passed (e.g. .../documents/...)
// 	// Need to go up one level from 'documents' if present.

// 	baseDir := filepath.Dir(kindlePath)
// 	if filepath.Base(baseDir) == "documents" {
// 		baseDir = filepath.Dir(baseDir)
// 	}

// 	thumbDir := filepath.Join(baseDir, "system", "thumbnails")
// 	if _, err := os.Stat(thumbDir); os.IsNotExist(err) {
// 		os.MkdirAll(thumbDir, 0755)
// 	}

// 	target := filepath.Join(thumbDir, "thumbnail_"+asin+"_EBOK_portrait.jpg")
// 	err := imaging.Save(c.image, target, imaging.JPEGQuality(85))
// 	if err != nil {
// 		return fmt.Errorf("failed to save cover to kindle: %w", err)
// 	}
// 	return nil
// }
