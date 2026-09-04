package turso

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
)

func testProjectRepository(t *testing.T) (*ProjectRepository, *domain.User) {
	t.Helper()

	db := testDatabase(t)

	users := NewUserRepository(db)
	projects := NewProjectRepository(db)

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

	return projects, user
}

func TestProjectRepositoryCreate(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		OwnerID:     user.ID,
		Title:       "Flowerpress",
		Slug:        "flowerpress",
		Description: "Personal publishing system.",
		Status:      domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if project.ID == 0 {
		t.Fatal("expected project ID")
	}

	if project.OwnerID != user.ID {
		t.Fatalf(
			"expected owner ID %d, got %d",
			user.ID,
			project.OwnerID,
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
}

func TestProjectRepositoryByID(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		OwnerID: user.ID,
		Title:   "Flowerpress",
		Slug:    "flowerpress",
		Status:  domain.ProjectStatusDraft,
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
}

func TestProjectRepositoryBySlug(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		OwnerID: user.ID,
		Title:   "Flowerpress",
		Slug:    "flowerpress",
		Status:  domain.ProjectStatusDraft,
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

func TestProjectRepositoryByIDNotFound(t *testing.T) {
	repo, _ := testProjectRepository(t)

	_, err := repo.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrPorjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryBySlugNotFound(t *testing.T) {
	repo, _ := testProjectRepository(t)

	_, err := repo.BySlug(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryCreateRejectDuplicateSlug(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	first := &domain.Project{
		OwnerID: user.ID,
		Title:   "First",
		Slug:    "flowerpress",
		Status:  domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		OwnerID: user.ID,
		Title:   "Second",
		Slug:    "flowerpress",
		Status:  domain.ProjectStatusDraft,
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first project: %v", err)
	}

	err := repo.Create(ctx, second)
	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf(
			"expected ErrSlugTaken, got %v",
			err,
		)
	}
}
