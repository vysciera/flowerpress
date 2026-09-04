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

func TestProjectSErviceCreateResolvesDuplicateSlugs(t *testing.T) {
	projects, user := testProjectService(t)
	ctx := context.Background()

	first, err := projects.Create(
		ctx,
		user.ID,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}

	second, err := projects.Create(
		ctx,
		user.ID,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}

	third, err := projects.Create(
		ctx,
		user.ID,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create third project: %v", err)
	}

	if first.Slug != "flowerpress" {
		t.Fatalf("expected first slug %q, got %q", "flowerpress", first.Slug)
	}

	if second.Slug != "flowerpress-2" {
		t.Fatalf("expected second slug %q, got %q", "flowerpress-2", second.Slug)
	}

	if third.Slug != "flowerpress-3" {
		t.Fatalf("expected third slug %q, got %q", "flowerpress-3", third.Slug)
	}
}

func TestProjectServiceUpdate(t *testing.T) {
	projects, user := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		user.ID,
		"Flowerpress",
		"Old description",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	originalSlug := project.Slug

	updated, err := projects.Update(
		ctx,
		user.ID,
		project.ID,
		"Flowerpress Archive",
		"New Description",
	)
	if err != nil {
		t.Fatalf("update project: %v", err)
	}

	if updated.Title != "Flowerpress Archive" {
		t.Fatalf("expected title %q, got %q", "Flowerpress Archive", updated.Title)
	}

	if updated.Description != "New Description" {
		t.Fatalf("expected description %q, got %q", "New Description", updated.Description)
	}

	if updated.Slug != originalSlug {
		t.Fatalf("expected slug %q to remain stable, got %q", originalSlug, updated.Slug)
	}
}

func TestProjectServiceUpdateRequiresTitle(t *testing.T) {
	projects, user := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		user.ID,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = projects.Update(
		ctx,
		user.ID,
		project.ID,
		"   ",
		"description",
	)

	if !errors.Is(err, ErrProjectTitleRequired) {
		t.Fatalf("expected ErrProjectTitleRequired, got %v", err)
	}
}

func TestProjectServiceUpdateRejectsDifferentOwner(t *testing.T) {
	db := testDatabase(t)

	users := turso.NewUserRepository(db)
	repo := turso.NewProjectRepository(db)
	projects := NewProjectService(repo)

	ctx := context.Background()

	owner := &domain.User{
		Username:     "flower",
		PasswordHash: "flowerhash",
	}

	other := &domain.User{
		Username:     "garden",
		PasswordHash: "gardenhash",
	}

	if err := users.Create(ctx, owner); err != nil {
		t.Fatalf("create owner: %v", err)
	}

	if err := users.Create(ctx, other); err != nil {
		t.Fatalf("create other user: %v", err)
	}

	project, err := projects.Create(
		ctx,
		owner.ID,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = projects.Update(
		ctx,
		other.ID,
		project.ID,
		"Changed",
		"",
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}
