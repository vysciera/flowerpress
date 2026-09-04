package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusPublished ProjectStatus = "published"
	ProjectStatusUnlisted  ProjectStatus = "unlisted"
	ProjectStatusArchived  ProjectStatus = "archived"
)

type Project struct {
	ID          int64 // Nine quintillion users on a single user application
	OwnerID     int64 // Same here
	Title       string
	Slug        string
	Description string
	Status      ProjectStatus

	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
