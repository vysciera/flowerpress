package turso

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
)

func testProjectRepository(t *testing.T) *ProjectRepository {
	t.Helper()

	return NewProjectRepository(
		testDatabase(t),
	)
}

func TestProjectRepositoryCreate(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:       "Flowerpress",
		Slug:        "flowerpress",
		Description: "Personal archive",
		Status:      domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if project.ID == 0 {
		t.Fatal("expected project ID")
	}

	if project.Title != "Flowerpress" {
		t.Fatalf(
			"expected title %q, got %q",
			"Flowerpress",
			project.Title,
		)
	}

	if project.Slug != "flowerpress" {
		t.Fatalf(
			"expected slug %q, got %q",
			"flowerpress",
			project.Slug,
		)
	}

	if project.Description != "Personal archive" {
		t.Fatalf(
			"expected description %q, got %q",
			"Personal archive",
			project.Description,
		)
	}

	if project.Status != domain.ProjectStatusDraft {
		t.Fatalf(
			"expected status %q, got %q",
			domain.ProjectStatusDraft,
			project.Status,
		)
	}

	if project.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if project.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}

	if project.PublishedAt != nil {
		t.Fatalf("expected nil PublishedAt, got %v", project.PublishedAt)
	}
}

func TestProjectRepositoryByID(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create projects: %v", err)
	}

	found, err := repo.ByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}

	if found.ID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			found.ID,
		)
	}

	if found.Title != "Flowerpress" {
		t.Fatalf(
			"expected title %q, got %q",
			"Flowerpress",
			found.Title,
		)
	}

	if found.Slug != project.Slug {
		t.Fatalf(
			"expected slug %q, got %q",
			project.Slug,
			found.Slug,
		)
	}
}

func TestProjectRepositoryByIDNotFound(t *testing.T) {
	repo := testProjectRepository(t)

	_, err := repo.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryBySlug(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	found, err := repo.BySlug(ctx, "flowerpress")
	if err != nil {
		t.Fatalf("find project: %v", err)
	}

	if found.ID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			found.ID,
		)
	}
}

func TestProjectRepositoryBySlugNotFound(t *testing.T) {
	repo := testProjectRepository(t)

	_, err := repo.BySlug(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepositoryCreateRejectDuplicateSlug(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	first := &domain.Project{
		Title:  "First",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		Title:  "Second",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first project: %v", err)
	}

	err := repo.Create(ctx, second)
	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf("expected ErrSlugTaken, got %v", err)
	}
}

func TestProjectRepositoryList(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	first := &domain.Project{
		Title:  "First",
		Slug:   "first",
		Status: domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		Title:  "Second",
		Slug:   "second",
		Status: domain.ProjectStatusDraft,
	}

	third := &domain.Project{
		Title:  "Misc",
		Slug:   "misc",
		Status: domain.ProjectStatusDraft,
	}

	for _, project := range []*domain.Project{first, second, third} {
		if err := repo.Create(ctx, project); err != nil {
			t.Fatalf("create project %q: %v", project.Title, err)
		}
	}

	found, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(found))
	}

	ids := make(map[int64]bool)

	for _, project := range found {
		ids[project.ID] = true
	}

	for _, project := range []*domain.Project{first, second, third} {
		if !ids[project.ID] {
			t.Fatalf("expected project ID %d in list", project.ID)
		}
	}
}

func TestProjectRepositoryListEmpty(t *testing.T) {
	repo := testProjectRepository(t)

	found, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(found) != 0 {
		t.Fatalf("expected empty list, got %d projects", len(found))
	}
}

func TestProjectRepositoryListPublished(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	draft := &domain.Project{
		Title:  "Draft",
		Slug:   "draft",
		Status: domain.ProjectStatusDraft,
	}

	published := &domain.Project{
		Title:  "Published",
		Slug:   "published",
		Status: domain.ProjectStatusPublished,
	}

	unlisted := &domain.Project{
		Title:  "Unlisted",
		Slug:   "unlisted",
		Status: domain.ProjectStatusUnlisted,
	}

	archived := &domain.Project{
		Title:  "Archived",
		Slug:   "archived",
		Status: domain.ProjectStatusArchived,
	}

	for _, project := range []*domain.Project{draft, published, unlisted, archived} {
		if err := repo.Create(ctx, project); err != nil {
			t.Fatalf(
				"create project %q: %v",
				project.Title,
				err,
			)
		}
	}

	found, err := repo.ListPublished(ctx)
	if err != nil {
		t.Fatalf("list published projects: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 project, got %d", len(found))
	}

	if found[0].ID != published.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			published.ID,
			found[0].ID,
		)
	}
}

func TestProjectRepositoryUpdate(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:       "Flowerpress",
		Slug:        "flowerpress",
		Description: "old description",
		Status:      domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	project.Title = "New Flowerpress"
	project.Slug = "new-flowerpress"
	project.Description = "NEW!!! DESCRIPTION!!!"
	project.Status = domain.ProjectStatusPublished

	if err := repo.Update(ctx, project); err != nil {
		t.Fatalf("update project: %v", err)
	}

	found, err := repo.ByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}

	if found.Title != "New Flowerpress" {
		t.Fatalf(
			"expected title %q, got %q",
			"New Flowerpress",
			found.Title,
		)
	}

	if found.Slug != "new-flowerpress" {
		t.Fatalf(
			"expected slug %q, got %q",
			"new-flowerpress",
			found.Slug,
		)
	}

	if found.Description != "NEW!!! DESCRIPTION!!!" {
		t.Fatalf(
			"expected updated description, got %q",
			found.Description,
		)
	}

	if found.Status != domain.ProjectStatusPublished {
		t.Fatalf(
			"expected project status %q, got %q",
			domain.ProjectStatusPublished,
			found.Status,
		)
	}
}

func TestProjectRepositoryUpdatePublishedAt(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	now := project.CreatedAt

	project.Status = domain.ProjectStatusPublished
	project.PublishedAt = &now

	if err := repo.Update(ctx, project); err != nil {
		t.Fatalf("update project: %v", err)
	}

	if project.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}

	if !project.PublishedAt.Equal(now) {
		t.Fatalf(
			"expected PublishedAt %v, got %v",
			now,
			project.PublishedAt,
		)
	}
}

func TestProjectRepositoryUpdateNotFound(t *testing.T) {
	repo := testProjectRepository(t)

	project := &domain.Project{
		ID:     999,
		Title:  "MISSING",
		Slug:   "missing",
		Status: domain.ProjectStatusDraft,
	}

	err := repo.Update(context.Background(), project)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryUpdateRejectsDuplicateSlug(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	first := &domain.Project{
		Title:  "First",
		Slug:   "first",
		Status: domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		Title:  "Second",
		Slug:   "second",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first project: %v", err)
	}

	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second project: %v", err)
	}

	second.Slug = first.Slug
	err := repo.Update(ctx, second)

	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf(
			"expected ErrSlugTaken, got %v",
			err,
		)
	}
}

func TestProjectRepositoryDelete(t *testing.T) {
	repo := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		Title:  "Flowerpress",
		Slug:   "flowerpress",
		Status: domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("Create project: %v", err)
	}

	if err := repo.Delete(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	_, err := repo.ByID(ctx, project.ID)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryDeleteNotFound(t *testing.T) {
	repo := testProjectRepository(t)

	err := repo.Delete(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}
