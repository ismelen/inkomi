package epub

import (
	"archive/zip"
	"fmt"
	"os"
	"sort"
)

func writeEpubZip(outputPath string, files map[string][]byte) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear %s: %w", outputPath, err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	mimeWriter, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		return err
	}
	if _, err := mimeWriter.Write(files["mimetype"]); err != nil {
		return err
	}

	names := make([]string, 0, len(files)-1)
	for name := range files {
		if name != "mimetype" {
			names = append(names, name)
		}
	}
	sort.Strings(names) // orden determinista, útil para reproducibilidad/tests

	for _, name := range names {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
		if err != nil {
			return err
		}
		if _, err := w.Write(files[name]); err != nil {
			return err
		}
	}

	return zw.Close()
}
