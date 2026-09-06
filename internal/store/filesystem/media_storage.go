package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"flowerpress/internal/service"
)

type MediaStorage struct {
	root string
}

func NewMediaStorage(root string) *MediaStorage {
	return &MediaStorage{
		root: root,
	}
}

var _ service.MediaStorage = (*MediaStorage)(nil)

func (s *MediaStorage) Put(ctx context.Context, key string, source io.Reader) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create media directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create media file: %w", err)
	}

	success := false

	defer func() {
		_ = file.Close()

		if !success {
			_ = os.Remove(path)
		}
	}()

	if _, err := copyWithContext(ctx, file, source); err != nil {
		return fmt.Errorf("write media file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close media file: %w", err)
	}

	success = true

	return nil
}

func (s *MediaStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open media file: %w", err)
	}

	return file, nil
}

func (s *MediaStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}

	err = os.Remove(path)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("delete media file: %w", err)
	}

	return nil
}

func (s *MediaStorage) pathForKey(key string) (string, error) {
	key = filepath.Clean(filepath.FromSlash(key))

	if key == "." || filepath.IsAbs(key) || key == ".." || strings.HasPrefix(key, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid media storage key")
	}

	return filepath.Join(s.root, key), nil
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buffer := make([]byte, 32*1024)

	var total int64

	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}

		n, readErr := src.Read(buffer)

		if n > 0 {
			written, writeErr := dst.Write(buffer[:n])

			total += int64(written)

			if writeErr != nil {
				return total, writeErr
			}

			if written != n {
				return total, io.ErrShortWrite
			}
		}

		if errors.Is(readErr, io.EOF) {
			return total, nil
		}

		if readErr != nil {
			return total, readErr
		}
	}
}
