package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/cloud"
)

var _ cloud.ICloudStorage = (*LocalStorage)(nil)

var (
	ErrInvalidKey = errors.New("invalid storage key")
	ErrTooLarge   = errors.New("file larger than allowed")
)

// LocalStorage keeps id/filename as root/id/filename on the server's disk. Files move in and out
// with hard links when the other side is on the same filesystem (the workers' temp dirs live on
// the data volume), so storing or fetching a file copies nothing.
type LocalStorage struct {
	root string
	tmp  string
}

// NewLocalStorage stores the files under root and writes partial uploads to tmp, which must be
// on the same filesystem so they can be renamed into place.
func NewLocalStorage(root, tmp string) (*LocalStorage, error) {
	for _, dir := range []string{root, tmp} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return &LocalStorage{root: root, tmp: tmp}, nil
}

// Root is the directory that holds the files, used to measure its free space.
func (s *LocalStorage) Root() string {
	return s.root
}

// Path returns where id/filename lives, rejecting anything that could leave the root.
func (s *LocalStorage) Path(id, filename string) (string, error) {
	for _, part := range []string{id, filename} {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, "/\\\x00") {
			return "", fmt.Errorf("%w: %q", ErrInvalidKey, id+"/"+filename)
		}
	}
	return filepath.Join(s.root, id, filename), nil
}

func (s *LocalStorage) Upload(ctx context.Context, id string, filename string, body io.Reader) error {
	_, err := s.Write(id, filename, body, -1)
	return err
}

// Write stores body as id/filename through a temp file, so readers never see a partial file.
// With limit >= 0, a body longer than limit bytes fails with ErrTooLarge and stores nothing.
func (s *LocalStorage) Write(id, filename string, body io.Reader, limit int64) (int64, error) {
	dst, err := s.Path(id, filename)
	if err != nil {
		return 0, err
	}
	tmp, err := os.CreateTemp(s.tmp, "upload-*")
	if err != nil {
		return 0, fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		// already gone when it was renamed into place
		_ = os.Remove(tmp.Name())
	}()

	reader := body
	if limit >= 0 {
		reader = io.LimitReader(body, limit+1)
	}
	n, err := io.Copy(tmp, reader)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("write %s: %w", dst, err), tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return 0, fmt.Errorf("close %s: %w", tmp.Name(), err)
	}
	if limit >= 0 && n > limit {
		return 0, ErrTooLarge
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return 0, fmt.Errorf("create %s: %w", filepath.Dir(dst), err)
	}
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return 0, fmt.Errorf("store %s: %w", dst, err)
	}
	return n, nil
}

func (s *LocalStorage) Download(ctx context.Context, id string, filename string) (io.ReadCloser, error) {
	path, err := s.Path(id, filename)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *LocalStorage) UploadFile(ctx context.Context, id string, filename string, path string) (int64, error) {
	dst, err := s.Path(id, filename)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return 0, fmt.Errorf("create %s: %w", filepath.Dir(dst), err)
	}
	if err := linkOrCopy(path, dst); err != nil {
		return 0, fmt.Errorf("store %s/%s: %w", id, filename, err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", dst, err)
	}
	return info.Size(), nil
}

func (s *LocalStorage) DownloadFile(ctx context.Context, id string, filename string, path string) error {
	src, err := s.Path(id, filename)
	if err != nil {
		return err
	}
	if err := linkOrCopy(src, path); err != nil {
		return fmt.Errorf("fetch %s/%s: %w", id, filename, err)
	}
	return nil
}

func (s *LocalStorage) Size(ctx context.Context, id string, filename string) (int64, error) {
	path, err := s.Path(id, filename)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Exists reports whether id/filename is stored.
func (s *LocalStorage) Exists(id, filename string) bool {
	_, err := s.Size(context.Background(), id, filename)
	return err == nil
}

// Delete removes id/filename and the id directory once it is empty; a missing file is not an error.
func (s *LocalStorage) Delete(ctx context.Context, id string, filename string) error {
	path, err := s.Path(id, filename)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// fails while the directory still holds other files, which is fine
	_ = os.Remove(filepath.Dir(path))
	return nil
}

// DeleteAll removes every file stored under id.
func (s *LocalStorage) DeleteAll(id string) error {
	if _, err := s.Path(id, "x"); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.root, id))
}

// Usage returns the bytes stored under each of ids.
func (s *LocalStorage) Usage(ids []string) int64 {
	var total int64
	for _, id := range ids {
		entries, err := os.ReadDir(filepath.Join(s.root, id))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if info, err := e.Info(); err == nil && info.Mode().IsRegular() {
				total += info.Size()
			}
		}
	}
	return total
}

// linkOrCopy makes dst the same file as src: a hard link when both are on the same filesystem,
// a copy otherwise. An existing dst is replaced.
func linkOrCopy(src, dst string) error {
	if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Link(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		return errors.Join(err, out.Close())
	}
	return out.Close()
}
