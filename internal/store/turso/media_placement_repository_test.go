package turso

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
)

func testMediaPlacementRepository(t *testing.T) (*MediaPlacementRepository, *ProjectRepository, *MediaAssetRepository) {
	t.Helper()

	db := testDatabase(t)

	return NewMediaPlacementRepository(db),
		NewProjectRepository(db),
		NewMediaAssetRepository(db)
}

func createTestProject(t *testing.T, repo *ProjectRepository, slug string) *domain.Project {
	t.Helper()

	project := &domain.Project{
		Title:  slug,
		Slug:   slug,
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(
		context.Background(),
		project,
	); err != nil {
		t.Fatalf("create test project: %v", err)
	}

	return project
}

func createTestAsset(t *testing.T, repo *MediaAssetRepository, key string, hash string) *domain.MediaAsset {
	t.Helper()

	asset := &domain.MediaAsset{
		StorageKey:   key,
		OriginalName: key,
		MIMEType:     "image/jpeg",
		SizeBytes:    100,
		SHA256:       hash,
	}

	if err := repo.Create(
		context.Background(),
		asset,
	); err != nil {
		t.Fatalf("create test asset: %v", err)
	}

	return asset
}

func TestMediaPlacementRepositoryCreate(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/flower.jpg",
		"flower-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
		Position:  2,
		Caption:   "Flower",
		AltText:   "A flower",
	}

	if err := placements.Create(ctx, placement); err != nil {
		t.Fatalf("create media placement: %v", err)
	}

	if placement.ID == 0 {
		t.Fatal("expected media placement ID")
	}

	if placement.AssetID != asset.ID {
		t.Fatalf(
			"expected asset ID %d, got %d",
			asset.ID,
			placement.AssetID,
		)
	}

	if placement.ProjectID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			placement.ProjectID,
		)
	}

	if placement.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if placement.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}
}

func TestMediaPlacementRepositoryByID(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/image.jpg",
		"image-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
	}

	if err := placements.Create(ctx, placement); err != nil {
		t.Fatalf("create placement: %v", err)
	}

	found, err := placements.ByID(ctx, placement.ID)
	if err != nil {
		t.Fatalf("find placement: %v", err)
	}

	if found.ID != placement.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			placement.ID,
			found.ID,
		)
	}
}

func TestMediaPlacementRepositoryByIDNotFound(t *testing.T) {
	placements, _, _ := testMediaPlacementRepository(t)

	_, err := placements.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrMediaPlacementNotFound) {
		t.Fatalf("expected ErrMediaPlacementNotFound, got %v", err)
	}
}

func TestMediaPlacementRepositoryListByProject(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	firstAsset := createTestAsset(
		t,
		assets,
		"media/first.jpg",
		"first-hash",
	)

	secondAsset := createTestAsset(
		t,
		assets,
		"media/second.jpg",
		"second-hash",
	)

	first := &domain.MediaPlacement{
		AssetID:   firstAsset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
		Position:  2,
	}

	second := &domain.MediaPlacement{
		AssetID:   secondAsset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
		Position:  1,
	}

	if err := placements.Create(ctx, first); err != nil {
		t.Fatalf("create first placement: %v", err)
	}

	if err := placements.Create(ctx, second); err != nil {
		t.Fatalf("create second placement: %v", err)
	}

	found, err := placements.ListByProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("list placements: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf(
			"expected 2 placements, got %d",
			len(found),
		)
	}

	if found[0].ID != second.ID {
		t.Fatalf("expected position 1 first")
	}

	if found[1].ID != first.ID {
		t.Fatalf("expected position 2 second")
	}
}

func TestMediaPlacementRepositoryRejectsSecondThumbnail(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	firstAsset := createTestAsset(
		t,
		assets,
		"media/first.jpg",
		"first-hash",
	)

	secondAsset := createTestAsset(
		t,
		assets,
		"media/second.jpg",
		"second-hash",
	)

	first := &domain.MediaPlacement{
		AssetID:   firstAsset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementThumbnail,
	}

	second := &domain.MediaPlacement{
		AssetID:   secondAsset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementThumbnail,
	}

	if err := placements.Create(ctx, first); err != nil {
		t.Fatalf("create first thumbnail: %v", err)
	}

	if err := placements.Create(ctx, second); err == nil {
		t.Fatal("expected second thumbnail to fail")
	}
}

func TestMediaPlacementRepositoryUpdate(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/image.jpg",
		"image-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
		Position:  0,
	}

	if err := placements.Create(
		ctx,
		placement,
	); err != nil {
		t.Fatalf("create placement: %v", err)
	}

	placement.Role = domain.MediaPlacementAttachment
	placement.Position = 3
	placement.Caption = "Updated caption"
	placement.AltText = "Updated alt text"

	if err := placements.Update(
		ctx,
		placement,
	); err != nil {
		t.Fatalf("update placement: %v", err)
	}

	if placement.Role != domain.MediaPlacementAttachment {
		t.Fatalf(
			"expected attachment role, got %q",
			placement.Role,
		)
	}

	if placement.Position != 3 {
		t.Fatalf(
			"expected position 3, got %d",
			placement.Position,
		)
	}

	if placement.Caption != "Updated caption" {
		t.Fatalf(
			"unexpected caption %q",
			placement.Caption,
		)
	}

	if placement.AltText != "Updated alt text" {
		t.Fatalf(
			"unexpected alt text %q",
			placement.AltText,
		)
	}
}

func TestMediaPlacementRepositoryDelete(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/image.jpg",
		"image-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
	}

	if err := placements.Create(ctx, placement); err != nil {
		t.Fatalf("create placement: %v", err)
	}

	if err := placements.Delete(ctx, placement.ID); err != nil {
		t.Fatalf("delete placement: %v", err)
	}

	_, err := placements.ByID(ctx, placement.ID)
	if !errors.Is(err, domain.ErrMediaPlacementNotFound) {
		t.Fatalf(
			"expected ErrMediaPlacementNotFound, got %v",
			err,
		)
	}
}

func TestMediaPlacementDeletedWithProject(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/image.jpg",
		"image-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
	}

	if err := placements.Create(ctx, placement); err != nil {
		t.Fatalf("create placement: %v", err)
	}

	if err := projects.Delete(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	_, err := placements.ByID(ctx, placement.ID)
	if !errors.Is(err, domain.ErrMediaPlacementNotFound) {
		t.Fatalf("expected placement to be deleted, got %v", err)
	}
}

func TestMediaPlacementDeletedWithAsset(t *testing.T) {
	placements, projects, assets := testMediaPlacementRepository(t)
	ctx := context.Background()

	project := createTestProject(
		t,
		projects,
		"flowerpress",
	)

	asset := createTestAsset(
		t,
		assets,
		"media/image.jpg",
		"image-hash",
	)

	placement := &domain.MediaPlacement{
		AssetID:   asset.ID,
		ProjectID: project.ID,
		Role:      domain.MediaPlacementContent,
	}

	if err := placements.Create(ctx, placement); err != nil {
		t.Fatalf("create placement: %v", err)
	}

	if err := assets.Delete(ctx, asset.ID); err != nil {
		t.Fatalf("delete asset: %v", err)
	}

	_, err := placements.ByID(ctx, placement.ID)
	if !errors.Is(err, domain.ErrMediaPlacementNotFound) {
		t.Fatalf("expected placement to be deleted, got %v", err)
	}
}
