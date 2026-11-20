package downloadbook

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func saveStream(stream io.Reader, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if _, err := io.Copy(file, stream); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", path, err), file.Close())
	}
	return file.Close()
}
