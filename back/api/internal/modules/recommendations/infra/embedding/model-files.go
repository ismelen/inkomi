package embedding

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxModelFileSize = 1 << 30

// tokenizerFiles are what the tokenizer of a Hugging Face model needs next to its ONNX file.
var tokenizerFiles = []string{"config.json", "tokenizer.json", "tokenizer_config.json", "special_tokens_map.json"}

// ModelFiles keeps the files of the embedding model in dir, downloading them from baseURL (a
// Hugging Face "resolve/main" url) the first time.
type ModelFiles struct {
	dir      string
	baseURL  string
	onnxFile string
	client   *http.Client

	mu sync.Mutex
}

func NewModelFiles(dir, baseURL, onnxFile string) *ModelFiles {
	return &ModelFiles{
		dir:      dir,
		baseURL:  strings.TrimRight(baseURL, "/"),
		onnxFile: onnxFile,
		client:   &http.Client{Timeout: 15 * time.Minute},
	}
}

// OnnxFilename is the name of the ONNX file inside the model directory.
func (m *ModelFiles) OnnxFilename() string {
	return path.Base(m.onnxFile)
}

// Ensure downloads the missing files and returns the model directory.
func (m *ModelFiles) Ensure(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	files := map[string]string{}
	for _, name := range tokenizerFiles {
		files[name] = name
	}
	files[m.onnxFile] = m.OnnxFilename()

	if err := os.MkdirAll(m.dir, 0o750); err != nil {
		return "", fmt.Errorf("create model dir: %w", err)
	}
	for remote, local := range files {
		target := filepath.Join(m.dir, local)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		log.Printf("recommendations: downloading %s/%s", m.baseURL, remote)
		if err := m.download(ctx, m.baseURL+"/"+remote, target); err != nil {
			return "", err
		}
	}
	return m.dir, nil
}

func (m *ModelFiles) download(ctx context.Context, url, target string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request %s: %w", url, err)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(m.dir, ".download-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(tmp.Name()))
		}
	}()
	written, err := io.Copy(tmp, io.LimitReader(resp.Body, maxModelFileSize+1))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	if written > maxModelFileSize {
		return fmt.Errorf("download %s: bigger than %d bytes", url, maxModelFileSize)
	}
	return os.Rename(tmp.Name(), target)
}
