package service

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
	"flowerpress/internal/store/turso"
)

func testProjectService(t *testing.T) (*ProjectService, *domain.User) {
	t.Helper()

	db := testDatabase(t)

	users := turso.NewUserRepository(db)
	projects := turso.NewProjectRepository(db)

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

	return NewProjectService(projects), user
}

func TestProjectServiceCreate(t *testing.T) {
	projects, user := testProjectService(t)

	project, err := projects.Create(
		context.Background(),
		user.ID,
		"Flower Press",
		"Personal archive",
	)

	if err != nil {
		t.Fatalf("Create project: %v", err)
	}

	if project.ID == 0 {
		t.Fatal("expected project ID")
	}

	if project.OwnerID != user.ID {
		t.Fatalf("expected owner ID %d, got %d", user.ID, project.OwnerID)
	}

	if project.Title != "Flower Press" {
		t.Fatalf("expected title %q, got %q", "Flower Press", project.Title)
	}

	if project.Slug != "flower-press" {
		t.Fatalf("expected slug %q, got %q", "flower-press", project.Slug)
	}

	if project.Status != domain.ProjectStatusDraft {
		t.Fatalf("expected draft status, got %q", project.Status)
	}
}

func TestProjectServiceCreateRequiresTitle(t *testing.T) {
	projects, user := testProjectService(t)

	_, err := projects.Create(
		context.Background(),
		user.ID,
		"   ",
		"description",
	)

	if !errors.Is(err, ErrProjectTitleRequired) {
		t.Fatalf("expected ErrProjectTitleRequired, got %v", err)
	}
}

func TestProjectServiceCreateNormalizesSlug(t *testing.T) {
	projects, user := testProjectService(t)

	project, err := projects.Create(
		context.Background(),
		user.ID,
		"  New      Origami:    2026!     ",
		"",
	)

	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if project.Slug != "new-origami-2026" {
		t.Fatalf(
			"expected slug %q, got %q",
			"new-origami-2026",
			project.Slug,
		)
	}
}
