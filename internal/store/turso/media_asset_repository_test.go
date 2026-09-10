package turso

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
)

func testMediaAssetRepository(t *testing.T) *MediaAssetRepository {
	t.Helper()

	return NewMediaAssetRepository(
		testDatabase(t),
	)
}

func TestMediaAssetRepositoryCreate(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	width := 1920
	height := 1080

	asset := &domain.MediaAsset{
		StorageKey:   "media/flower.jpg",
		OriginalName: "flower.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    123456,
		SHA256:       "abc123",
		Width:        &width,
		Height:       &height,
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	if asset.ID == 0 {
		t.Fatal("expected media asset ID")
	}

	if asset.StorageKey != "media/flower.jpg" {
		t.Fatalf(
			"expected storage key %q, got %q",
			"media/flower.jpg",
			asset.StorageKey,
		)
	}

	if asset.OriginalName != "flower.jpg" {
		t.Fatalf(
			"expected original name %q, got %q",
			"flower.jpg",
			asset.OriginalName,
		)
	}

	if asset.MIMEType != "image/jpeg" {
		t.Fatalf(
			"expected MIME type %q, got %q",
			"image/jpeg",
			asset.MIMEType,
		)
	}

	if asset.SizeBytes != 123456 {
		t.Fatalf("expected size 123456, got %d", asset.SizeBytes)
	}

	if asset.SHA256 != "abc123" {
		t.Fatalf(
			"expected SHA256 %q, got %q",
			"abc123",
			asset.SHA256,
		)
	}

	if asset.Width == nil || *asset.Width != 1920 {
		t.Fatalf("expected width 1920, got %v", asset.Width)
	}

	if asset.Height == nil || *asset.Height != 1080 {
		t.Fatalf("expected height 1080, got %v", asset.Height)
	}

	if asset.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if asset.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}
}

func TestMediaAssetRepositoryCreateWithoutDimensions(t *testing.T) {
	repo := testMediaAssetRepository(t)

	asset := &domain.MediaAsset{
		StorageKey:   "01JMEDIA/document.pdf",
		OriginalName: "document.pdf",
		MIMEType:     "application/pdf",
		SizeBytes:    5000,
		SHA256:       "pdfhash",
	}

	if err := repo.Create(
		context.Background(),
		asset,
	); err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	if asset.Width != nil {
		t.Fatalf("expected nil width, got %v", asset.Width)
	}

	if asset.Height != nil {
		t.Fatalf("expected nil height, got %v", asset.Height)
	}
}

func TestMediaAssetRepositoryRejectsDuplicateSHA256(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	first := &domain.MediaAsset{
		StorageKey:   "media/first.jpg",
		OriginalName: "first.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "same-hash",
	}

	second := &domain.MediaAsset{
		StorageKey:   "media/second.jpg",
		OriginalName: "second.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "same-hash",
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first asset: %v", err)
	}

	if err := repo.Create(ctx, second); err == nil {
		t.Fatal("expected duplicate SHA256 to fail")
	}
}

func TestMediaAssetRepositoryRejectsDuplicateStorageKey(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	first := &domain.MediaAsset{
		StorageKey:   "media/shared.jpg",
		OriginalName: "first.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "first-hash",
	}

	second := &domain.MediaAsset{
		StorageKey:   "media/shared.jpg",
		OriginalName: "second.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    200,
		SHA256:       "second-hash",
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first asset: %v", err)
	}

	if err := repo.Create(ctx, second); err == nil {
		t.Fatal("expected duplicate storage key to fail")
	}
}

func TestMediaAssetRepositoryByID(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		StorageKey:   "media/image.jpg",
		OriginalName: "image.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "abc123",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, err := repo.ByID(ctx, asset.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}

	if found.ID != asset.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			asset.ID,
			found.ID,
		)
	}
}

func TestMediaAssetRepositoryByIDNotFound(t *testing.T) {
	repo := testMediaAssetRepository(t)

	_, err := repo.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}

func TestMediaAssetRepositoryByStorageKey(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		StorageKey:   "media/unique.jpg",
		OriginalName: "unique.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "uniquehash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, err := repo.ByStorageKey(ctx, "media/unique.jpg")
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}

	if found.ID != asset.ID {
		t.Fatalf("expected ID %d, got %d", asset.ID, found.ID)
	}
}

func TestMediaAssetRepositoryByStorageKeyNotFound(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	_, err := repo.ByStorageKey(ctx, "media/missing.jpg")
	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}

func TestMediaAssetRepositoryBySHA256(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		StorageKey:   "media/dedup.jpg",
		OriginalName: "dedup.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "same-content-hash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, err := repo.BySHA256(ctx, "same-content-hash")
	if err != nil {
		t.Fatalf("find asset by hash: %v", err)
	}

	if found.ID != asset.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			asset.ID,
			found.ID,
		)
	}
}

func TestMediaAssetRepositoryBySHA256NotFound(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	_, err := repo.BySHA256(ctx, "missing-hash")

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}

func TestMediaAssetRepositoryList(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	first := &domain.MediaAsset{
		StorageKey:   "media/first.jpg",
		OriginalName: "first.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "first-hash",
	}

	second := &domain.MediaAsset{
		StorageKey:   "media/second.pdf",
		OriginalName: "second.pdf",
		MIMEType:     "application/pdf",
		SizeBytes:    200,
		SHA256:       "second-hash",
	}

	third := &domain.MediaAsset{
		StorageKey:   "media/third.zip",
		OriginalName: "third.zip",
		MIMEType:     "application/zip",
		SizeBytes:    300,
		SHA256:       "third-hash",
	}

	for _, asset := range []*domain.MediaAsset{first, second, third} {
		if err := repo.Create(ctx, asset); err != nil {
			t.Fatalf(
				"create asset %q: %v",
				asset.OriginalName,
				err,
			)
		}
	}

	found, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list media assets: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 media assets, got %d", len(found))
	}

	ids := make(map[int64]bool)

	for _, asset := range found {
		ids[asset.ID] = true
	}

	for _, asset := range []*domain.MediaAsset{first, second, third} {
		if !ids[asset.ID] {
			t.Fatalf("expected media asset ID %d in list", asset.ID)
		}
	}
}

func TestMediaAssetRepositoryListEmpty(
	t *testing.T,
) {
	repo := testMediaAssetRepository(t)

	found, err := repo.List(context.Background())

	if err != nil {
		t.Fatalf("list media assets: %v", err)
	}

	if len(found) != 0 {
		t.Fatalf("expected empty list, got %d assets", len(found))
	}
}

func TestMediaAssetRepositoryUpdate(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		StorageKey:   "media/original.jpg",
		OriginalName: "original.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "original-hash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	width := 800
	height := 600

	asset.StorageKey = "media/renamed.jpg"
	asset.OriginalName = "renamed.jpg"
	asset.MIMEType = "image/png"
	asset.SizeBytes = 200
	asset.SHA256 = "new-hash"
	asset.Width = &width
	asset.Height = &height

	if err := repo.Update(ctx, asset); err != nil {
		t.Fatalf("update media asset: %v", err)
	}

	if asset.StorageKey != "media/renamed.jpg" {
		t.Fatalf("unexpected storage key %q", asset.StorageKey)
	}

	if asset.OriginalName != "renamed.jpg" {
		t.Fatalf("unexpected original name %q", asset.OriginalName)
	}

	if asset.MIMEType != "image/png" {
		t.Fatalf("unexpected MIME type %q", asset.MIMEType)
	}

	if asset.SizeBytes != 200 {
		t.Fatalf("expected size 200, got %d", asset.SizeBytes)
	}

	if asset.SHA256 != "new-hash" {
		t.Fatalf("unexpected SHA256 %q", asset.SHA256)
	}

	if asset.Width == nil || *asset.Width != 800 {
		t.Fatalf("expected width 800, got %v", asset.Width)
	}

	if asset.Height == nil || *asset.Height != 600 {
		t.Fatalf("expected height 600, got %v", asset.Height)
	}
}

func TestMediaAssetRepositoryUpdateNotFound(t *testing.T) {
	repo := testMediaAssetRepository(t)

	asset := &domain.MediaAsset{
		ID:           999,
		StorageKey:   "media/missing.jpg",
		OriginalName: "missing.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "missing-hash",
	}

	err := repo.Update(context.Background(), asset)

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}

func TestMediaAssetRepositoryDelete(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		StorageKey:   "media/delete.jpg",
		OriginalName: "delete.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "delete-hash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	if err := repo.Delete(ctx, asset.ID); err != nil {
		t.Fatalf("delete media asset: %v", err)
	}

	_, err := repo.ByID(ctx, asset.ID)
	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}

func TestMediaAssetRepositoryDeleteNotFound(t *testing.T) {
	repo := testMediaAssetRepository(t)
	ctx := context.Background()

	err := repo.Delete(ctx, 999)
	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}
