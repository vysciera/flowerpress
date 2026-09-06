package filesystem

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestMediaStoragePutAndOpen(t *testing.T) {
	storage := NewMediaStorage(
		t.TempDir(),
	)

	ctx := context.Background()

	err := storage.Put(
		ctx,
		"ab/cd/file.txt",
		strings.NewReader("flowerpress"),
	)
	if err != nil {
		t.Fatalf("put file: %v", err)
	}

	file, err := storage.Open(ctx, "ab/cd/file.txt")
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if string(data) != "flowerpress" {
		t.Fatalf(
			"expected %q, got %q",
			"flowerpress",
			string(data),
		)
	}
}

func TestMediaStoragePutRejectsExistingKey(t *testing.T) {
	storage := NewMediaStorage(
		t.TempDir(),
	)

	ctx := context.Background()

	if err := storage.Put(
		ctx,
		"file.txt",
		strings.NewReader("first"),
	); err != nil {
		t.Fatalf("put first file: %v", err)
	}

	err := storage.Put(
		ctx,
		"file.txt",
		strings.NewReader("second"),
	)

	if err == nil {
		t.Fatal("expected existing key to fail")
	}
}

func TestMediaStorageDelete(t *testing.T) {
	storage := NewMediaStorage(
		t.TempDir(),
	)

	ctx := context.Background()

	if err := storage.Put(
		ctx,
		"file.txt",
		strings.NewReader("flowerpress"),
	); err != nil {
		t.Fatalf("put file: %v", err)
	}

	if err := storage.Delete(ctx, "file.txt"); err != nil {
		t.Fatalf("delete file: %v", err)
	}

	_, err := storage.Open(ctx, "file.txt")
	if err == nil {
		t.Fatal("expected deleted file to be missing")
	}
}

func TestMediaStorageRejectsPathTraversal(t *testing.T) {
	storage := NewMediaStorage(
		t.TempDir(),
	)

	err := storage.Put(
		context.Background(),
		"../../outside.txt",
		strings.NewReader("nope"),
	)

	if err == nil {
		t.Fatal(
			"expected path traversal to fail",
		)
	}
}
