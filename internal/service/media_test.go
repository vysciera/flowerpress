package service

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
	"flowerpress/internal/store/turso"
)

func testMediaService(t *testing.T) *MediaService {
	t.Helper()

	db := testDatabase(t)

	return NewMediaService(
		turso.NewMediaAssetRepository(db),
		turso.NewMediaPlacementRepository(db),
		turso.NewProjectRepository(db),
	)
}

func TestMediaServiceRegisterAsset(t *testing.T) {
	media := testMediaService(t)

	width := 1920
	height := 1080

	asset, err := media.RegisterAsset(
		context.Background(),
		"media/flower.jpg",
		"flower.jpg",
		"image/jpeg",
		123456,
		"flower-hash",
		&width,
		&height,
	)
	if err != nil {
		t.Fatalf("register asset: %v", err)
	}

	if asset.ID == 0 {
		t.Fatal("expected asset ID")
	}

	if asset.SHA256 != "flower-hash" {
		t.Fatalf(
			"expected SHA256 %q, got %q",
			"flower-hash",
			asset.SHA256,
		)
	}
}

func TestMediaServiceRegisterAssetDeduplicatesSHA256(t *testing.T) {
	media := testMediaService(t)
	ctx := context.Background()

	first, err := media.RegisterAsset(
		ctx,
		"media/first.jpg",
		"first.jpg",
		"image/jpeg",
		100,
		"same-hash",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("register first asset: %v", err)
	}

	second, err := media.RegisterAsset(
		ctx,
		"media/second.jpg",
		"second.jpg",
		"image/jpeg",
		100,
		"same-hash",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("register second asset: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf(
			"expected same asset ID, got %d and %d",
			first.ID,
			second.ID,
		)
	}

	found, err := media.ListAssets(ctx)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(found))
	}
}

func TestMediaServiceRegisterAssetRequiresSHA256(t *testing.T) {
	media := testMediaService(t)

	_, err := media.RegisterAsset(
		context.Background(),
		"media/image.jpg",
		"image.jpg",
		"image/jpeg",
		100,
		"",
		nil,
		nil,
	)

	if !errors.Is(err, ErrMediaSHA256Required) {
		t.Fatalf("expected ErrMediaSHA256Required, got %v", err)
	}
}

func TestMediaServicePlaceAsset(t *testing.T) {
	db := testDatabase(t)

	projectRepo := turso.NewProjectRepository(db)
	assetRepo := turso.NewMediaAssetRepository(db)
	placementRepo := turso.NewMediaPlacementRepository(db)

	media := NewMediaService(
		assetRepo,
		placementRepo,
		projectRepo,
	)

	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := projectRepo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	asset, err := media.RegisterAsset(
		ctx,
		"media/image.jpg",
		"image.jpg",
		"image/jpeg",
		100,
		"image-hash",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("register asset: %v", err)
	}

	placement, err := media.PlaceAsset(
		ctx,
		project.ID,
		asset.ID,
		domain.MediaPlacementContent,
		0,
		"Caption",
		"Alternative text",
	)
	if err != nil {
		t.Fatalf("place asset: %v", err)
	}

	if placement.ProjectID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			placement.ProjectID,
		)
	}

	if placement.AssetID != asset.ID {
		t.Fatalf(
			"expected asset ID %d, got %d",
			asset.ID,
			placement.AssetID,
		)
	}
}

func TestMediaServicePlaceAssetRequiresProject(t *testing.T) {
	media := testMediaService(t)
	ctx := context.Background()

	asset, err := media.RegisterAsset(
		ctx,
		"media/image.jpg",
		"image.jpg",
		"image/jpeg",
		100,
		"image-hash",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("register asset: %v", err)
	}

	_, err = media.PlaceAsset(
		ctx,
		999,
		asset.ID,
		domain.MediaPlacementContent,
		0,
		"",
		"",
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestMediaServicePlaceAssetRequiresAsset(t *testing.T) {
	db := testDatabase(t)

	projectRepo := turso.NewProjectRepository(db)

	media := NewMediaService(
		turso.NewMediaAssetRepository(db),
		turso.NewMediaPlacementRepository(db),
		projectRepo,
	)

	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := projectRepo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err := media.PlaceAsset(
		ctx,
		project.ID,
		999,
		domain.MediaPlacementContent,
		0,
		"",
		"",
	)

	if !errors.Is(err, domain.ErrMediaAssetNotFound) {
		t.Fatalf("expected ErrMediaAssetNotFound, got %v", err)
	}
}
