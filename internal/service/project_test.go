package service

import (
	"context"
	"errors"
	"testing"

	"flowerpress/internal/domain"
	"flowerpress/internal/store/turso"
)

func testProjectService(t *testing.T) *ProjectService {
	t.Helper()

	db := testDatabase(t)
	repo := turso.NewProjectRepository(db)

	return NewProjectService(repo)
}

func TestProjectServiceCreate(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flower Press",
		"Personal archive",
	)

	if err != nil {
		t.Fatalf("Create project: %v", err)
	}

	if project.ID == 0 {
		t.Fatal("expected project ID")
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

	if project.PublishedAt != nil {
		t.Fatalf("expected nil PublishedAt, got %v", project.PublishedAt)
	}
}

func TestProjectServiceCreateRequiresTitle(t *testing.T) {
	projects := testProjectService(t)

	_, err := projects.Create(
		context.Background(),
		"   ",
		"description",
	)

	if !errors.Is(err, ErrProjectTitleRequired) {
		t.Fatalf("expected ErrProjectTitleRequired, got %v", err)
	}
}

func TestProjectServiceCreateNormalizesSlug(t *testing.T) {
	projects := testProjectService(t)

	project, err := projects.Create(
		context.Background(),
		"  New   Origami: 2026!  ",
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

func TestProjectServiceCreateSequencesDuplicateSlugs(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	first, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}

	second, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}

	third, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create third project: %v", err)
	}

	if first.Slug != "flowerpress" {
		t.Fatalf(
			"expected first slug %q, got %q",
			"flowerpress",
			first.Slug,
		)
	}

	if second.Slug != "flowerpress-2" {
		t.Fatalf(
			"expected second slug %q, got %q",
			"flowerpress-2",
			second.Slug,
		)
	}

	if third.Slug != "flowerpress-3" {
		t.Fatalf(
			"expected third slug %q, got %q",
			"flowerpress-3",
			third.Slug,
		)
	}
}

func TestProjectServiceUpdate(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"Old description",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	originalSlug := project.Slug

	updated, err := projects.Update(
		ctx,
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
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = projects.Update(
		ctx,
		project.ID,
		"   ",
		"description",
	)

	if !errors.Is(err, ErrProjectTitleRequired) {
		t.Fatalf("expected ErrProjectTitleRequired, got %v", err)
	}
}

func TestProjectServiceUpdateNotFound(t *testing.T) {
	projects := testProjectService(t)

	_, err := projects.Update(
		context.Background(),
		999,
		"Missing",
		"",
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServicePublish(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	published, err := projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("publish project: %v", err)
	}

	if published.Status != domain.ProjectStatusPublished {
		t.Fatalf("expected status %q, got %q", domain.ProjectStatusPublished, published.Status)
	}

	if published.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}
}

func TestProjectServicePublishIsIdempotent(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	first, err := projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("first publish: %v", err)
	}

	second, err := projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("second publish: %v", err)
	}

	if first.PublishedAt == nil || second.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}

	if !first.PublishedAt.Equal(*second.PublishedAt) {
		t.Fatalf(
			"expected PublishedAt to remain unchanged: %v != %v",
			first.PublishedAt,
			second.PublishedAt,
		)
	}
}

func TestProjectServiceUnpublish(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("publish project: %v", err)
	}

	publishedAt := project.PublishedAt

	project, err = projects.Unpublish(ctx, project.ID)
	if err != nil {
		t.Fatalf("unpublish project: %v", err)
	}

	if project.Status != domain.ProjectStatusDraft {
		t.Fatalf("expected status %q, got %q", domain.ProjectStatusDraft, project.Status)
	}

	if project.PublishedAt == nil {
		t.Fatal("expected PublishedAt to be preserved")
	}

	if !project.PublishedAt.Equal(*publishedAt) {
		t.Fatal("expected original PublishedAt to be preserved")
	}
}

func TestProjectServiceUnlist(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Unlist(ctx, project.ID)
	if err != nil {
		t.Fatalf("unlist project: %v", err)
	}

	if project.Status != domain.ProjectStatusUnlisted {
		t.Fatalf("expected status %q, got %q", domain.ProjectStatusUnlisted, project.Status)
	}

	if project.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}
}

func TestProjectServiceUnlistPreservesPublishedAt(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create proejct: %v", err)
	}

	project, err = projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("publish project: %v", err)
	}

	publishedAt := *project.PublishedAt

	project, err = projects.Unlist(ctx, project.ID)
	if err != nil {
		t.Fatalf("unlist project: %v", err)
	}

	if project.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}

	if !project.PublishedAt.Equal(publishedAt) {
		t.Fatal("expected PublishedAt to be preserved")
	}
}

func TestProjectServiceArchive(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("Create project: %v", err)
	}

	project, err = projects.Archive(ctx, project.ID)
	if err != nil {
		t.Fatalf("archive project: %v", err)
	}

	if project.Status != domain.ProjectStatusArchived {
		t.Fatalf("expected status %q, got %q", domain.ProjectStatusArchived, project.Status)
	}
}

func TestProjectServiceCannotPublishArchivedProject(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Archive(ctx, project.ID)
	if err != nil {
		t.Fatalf("archive project: %v", err)
	}

	_, err = projects.Publish(ctx, project.ID)
	if !errors.Is(err, ErrInvalidProjectTransition) {
		t.Fatalf("expected ErrInvalidProjectTransition, got %v", err)
	}
}

func TestProjectServiceCannotUnpublishArchivedProject(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := projects.Archive(ctx, project.ID); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	_, err = projects.Unpublish(ctx, project.ID)
	if !errors.Is(err, ErrInvalidProjectTransition) {
		t.Fatalf(
			"expected ErrInvalidProjectTransition, got %v",
			err,
		)
	}
}

func TestProjectServiceArchivePreservesPublishedAt(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("publish project: %v", err)
	}

	publishedAt := *project.PublishedAt

	project, err = projects.Archive(ctx, project.ID)
	if err != nil {
		t.Fatalf("archive project: %v", err)
	}

	if project.PublishedAt == nil {
		t.Fatal("expected PublishedAt to be preserved")
	}

	if !project.PublishedAt.Equal(publishedAt) {
		t.Fatal("expected PublishedAt to remain unchanged")
	}
}

func TestProjectServiceByPublicSlugPublished(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Publish(ctx, project.ID)
	if err != nil {
		t.Fatalf("publish project: %v", err)
	}

	found, err := projects.ByPublicSlug(ctx, project.Slug)
	if err != nil {
		t.Fatalf("find public project: %v", err)
	}

	if found.ID != project.ID {
		t.Fatalf("expected project ID %d, got %d", project.ID, found.ID)
	}
}

func TestProjectServiceByPublicSlugUnlisted(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Unlist(ctx, project.ID)
	if err != nil {
		t.Fatalf("unlist project: %v", err)
	}

	found, err := projects.ByPublicSlug(ctx, project.Slug)
	if err != nil {
		t.Fatalf("find unlisted project: %v", err)
	}

	if found.ID != project.ID {
		t.Fatalf("expected project ID %d, got %d", project.ID, found.ID)
	}
}

func TestProjectServiceByPublicSlugHidesDraft(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = projects.ByPublicSlug(ctx, project.Slug)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceByPublicSlugHidesArchived(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := projects.Archive(ctx, project.ID); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	_, err = projects.ByPublicSlug(ctx, project.Slug)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceListPublic(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	first, err := projects.Create(
		ctx,
		"First",
		"",
	)
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}

	second, err := projects.Create(
		ctx,
		"Second",
		"",
	)
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}

	third, err := projects.Create(
		ctx,
		"Third",
		"",
	)
	if err != nil {
		t.Fatalf("create third project: %v", err)
	}

	if _, err := projects.Publish(
		ctx,
		first.ID,
	); err != nil {
		t.Fatalf("publish first project: %v", err)
	}

	if _, err := projects.Unlist(
		ctx,
		second.ID,
	); err != nil {
		t.Fatalf("unlist second project: %v", err)
	}

	if _, err := projects.Archive(
		ctx,
		third.ID,
	); err != nil {
		t.Fatalf("archive third project: %v", err)
	}

	found, err := projects.ListPublic(ctx)
	if err != nil {
		t.Fatalf("list public projects: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 public project, got %d", len(found))
	}

	if found[0].ID != first.ID {
		t.Fatalf("expected project ID %d, got %d", first.ID, found[0].ID)
	}
}

func TestProjectServiceDelete(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := projects.Delete(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	_, err = projects.ByID(ctx, project.ID)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceDeleteNotFound(t *testing.T) {
	projects := testProjectService(t)

	err := projects.Delete(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceCannotUnlistArchivedProject(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	project, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	project, err = projects.Archive(ctx, project.ID)
	if err != nil {
		t.Fatalf("archive project: %v", err)
	}

	_, err = projects.Unlist(ctx, project.ID)
	if !errors.Is(err, ErrInvalidProjectTransition) {
		t.Fatalf("expected ErrInvalidProjectTransition, got %v", err)
	}
}

func TestProjectServiceByID(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	created, err := projects.Create(
		ctx,
		"Flowerpress",
		"",
	)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	found, err := projects.ByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find projects: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected project ID %d, got %d", created.ID, found.ID)
	}
}

func TestProjectServiceByIDNotFound(t *testing.T) {
	projects := testProjectService(t)

	_, err := projects.ByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectServiceList(t *testing.T) {
	projects := testProjectService(t)
	ctx := context.Background()

	for _, title := range []string{"First", "Second", "Third"} {
		if _, err := projects.Create(
			ctx,
			title,
			"",
		); err != nil {
			t.Fatalf("create project %q: %v", title, err)
		}
	}

	found, err := projects.List(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(found))
	}
}
