package turso

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
)

func testMediaAssetRepository(
	t *testing.T,
) (*MediaAssetRepository, *domain.User) {
	t.Helper()

	db := testDatabase(t)

	users := NewUserRepository(db)
	assets := NewMediaAssetRepository(db)

	user := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	if err := users.Create(
		context.Background(),
		user,
	); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return assets, user
}

func TestMediaAssetRepositoryCreate(t *testing.T) {
	repo, user := testMediaAssetRepository(t)
	ctx := context.Background()

	width := 1920
	height := 1080

	asset := &domain.MediaAsset{
		OwnerID:      user.ID,
		StorageKey:   "01JMEDIA/image.jpg",
		OriginalName: "image.jpg",
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

	if asset.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if asset.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}

	if asset.Width == nil || *asset.Width != 1920 {
		t.Fatalf("expected width 1920, got %v", asset.Width)
	}

	if asset.Height == nil || *asset.Height != 1080 {
		t.Fatalf("expected height 1080, got %v", asset.Height)
	}
}

func TestMediaAssetRepositoryCreateWithoutDimensions(
	t *testing.T,
) {
	repo, user := testMediaAssetRepository(t)

	asset := &domain.MediaAsset{
		OwnerID:      user.ID,
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

func TestMediaAssetRepositoryByID(t *testing.T) {
	repo, user := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		OwnerID:      user.ID,
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
	repo, _ := testMediaAssetRepository(t)

	_, err := repo.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf(
			"expected ErrMediaAssetNotFound, got %v",
			err,
		)
	}
}

func TestMediaAssetRepositoryByStorageKey(t *testing.T) {
	repo, user := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		OwnerID:      user.ID,
		StorageKey:   "media/unique.jpg",
		OriginalName: "unique.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "uniquehash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, err := repo.ByStorageKey(
		ctx,
		"media/unique.jpg",
	)
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

func TestMediaAssetRepositoryBySHA256(t *testing.T) {
	repo, user := testMediaAssetRepository(t)
	ctx := context.Background()

	asset := &domain.MediaAsset{
		OwnerID:      user.ID,
		StorageKey:   "media/dedup.jpg",
		OriginalName: "dedup.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "same-content-hash",
	}

	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, err := repo.BySHA256(
		ctx,
		user.ID,
		"same-content-hash",
	)
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

func TestMediaAssetRepositoryBySHA256ScopesToOwner(
	t *testing.T,
) {
	db := testDatabase(t)

	users := NewUserRepository(db)
	assets := NewMediaAssetRepository(db)

	ctx := context.Background()

	first := &domain.User{
		Username:     "flower",
		PasswordHash: "hash",
	}

	second := &domain.User{
		Username:     "garden",
		PasswordHash: "hash",
	}

	if err := users.Create(ctx, first); err != nil {
		t.Fatalf("create first user: %v", err)
	}

	if err := users.Create(ctx, second); err != nil {
		t.Fatalf("create second user: %v", err)
	}

	asset := &domain.MediaAsset{
		OwnerID:      first.ID,
		StorageKey:   "media/owned.jpg",
		OriginalName: "owned.jpg",
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       "shared-hash",
	}

	if err := assets.Create(ctx, asset); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	_, err := assets.BySHA256(
		ctx,
		second.ID,
		"shared-hash",
	)

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf(
			"expected ErrMediaAssetNotFound, got %v",
			err,
		)
	}
}
