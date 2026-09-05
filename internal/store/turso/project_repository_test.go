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

func TestProjectRepositoryListByOwner(t *testing.T) {
	db := testDatabase(t)

	users := NewUserRepository(db)
	projects := NewProjectRepository(db)

	ctx := context.Background()

	firstOwner := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	secondOwner := &domain.User{
		Username:     "garden",
		PasswordHash: "gardenhash",
	}

	if err := users.Create(ctx, firstOwner); err != nil {
		t.Fatalf("create first owner: %v", err)
	}

	if err := users.Create(ctx, secondOwner); err != nil {
		t.Fatalf("create second owner: %v", err)
	}

	first := &domain.Project{
		OwnerID: firstOwner.ID,
		Title:   "First",
		Slug:    "first",
		Status:  domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		OwnerID: firstOwner.ID,
		Title:   "Second",
		Slug:    "second",
		Status:  domain.ProjectStatusDraft,
	}

	other := &domain.Project{
		OwnerID: secondOwner.ID,
		Title:   "Misc",
		Slug:    "misc",
		Status:  domain.ProjectStatusDraft,
	}

	for _, project := range []*domain.Project{first, second, other} {
		if err := projects.Create(ctx, project); err != nil {
			t.Fatalf("create project: %v", err)
		}
	}

	found, err := projects.ListByOwner(ctx, firstOwner.ID)

	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf(
			"expected 2 projects, got %d",
			len(found),
		)
	}

	for _, project := range found {
		if project.OwnerID != firstOwner.ID {
			t.Fatalf(
				"expected owner ID %d, got %d",
				firstOwner.ID,
				project.OwnerID,
			)
		}
	}
}

func TestProjectRepositoryupdate(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		OwnerID:     user.ID,
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
}

func TestProjectRepositoryUpdateNotFound(t *testing.T) {
	repo, user := testProjectRepository(t)

	project := &domain.Project{
		ID:      999,
		OwnerID: user.ID,
		Title:   "MISSING",
		Slug:    "missing",
		Status:  domain.ProjectStatusDraft,
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
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	first := &domain.Project{
		OwnerID: user.ID,
		Title:   "First",
		Slug:    "first",
		Status:  domain.ProjectStatusDraft,
	}

	second := &domain.Project{
		OwnerID: user.ID,
		Title:   "Second",
		Slug:    "second",
		Status:  domain.ProjectStatusDraft,
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
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	project := &domain.Project{
		OwnerID: user.ID,
		Title:   "Flowerpress",
		Slug:    "flowerpress",
		Status:  domain.ProjectStatusDraft,
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
	repo, _ := testProjectRepository(t)

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

func TestProjectRepositoryListPublished(t *testing.T) {
	repo, user := testProjectRepository(t)
	ctx := context.Background()

	draft := &domain.Project{
		OwnerID: user.ID,
		Title:   "Draft",
		Slug:    "draft",
		Status:  domain.ProjectStatusDraft,
	}

	published := &domain.Project{
		OwnerID: user.ID,
		Title:   "Published",
		Slug:    "published",
		Status:  domain.ProjectStatusPublished,
	}

	unlisted := &domain.Project{
		OwnerID: user.ID,
		Title:   "Unlisted",
		Slug:    "unlisted",
		Status:  domain.ProjectStatusUnlisted,
	}

	archived := &domain.Project{
		OwnerID: user.ID,
		Title:   "Archived",
		Slug:    "archived",
		Status:  domain.ProjectStatusArchived,
	}

	for _, project := range []*domain.Project{
		draft,
		published,
		unlisted,
		archived,
	} {
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
		t.Fatalf(
			"expected 1 project, got %d",
			len(found),
		)
	}

	if found[0].ID != published.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			published.ID,
			found[0].ID,
		)
	}
}
