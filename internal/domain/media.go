package domain

import "time"

type MediaAsset struct {
	ID int64

	StorageKey   string
	OriginalName string
	MIMEType     string
	SizeBytes    int64
	SHA256       string

	Width  *int
	Height *int

	CreatedAt time.Time
	UpdatedAt time.Time
}

type MediaPlacementRole string

const (
	MediaPlacementThumbnail  MediaPlacementRole = "thumbnail"
	MediaPlacementContent    MediaPlacementRole = "content"
	MediaPlacementAttachment MediaPlacementRole = "attachment"
)

type MediaPlacement struct {
	ID int64

	AssetID   int64
	ProjectID int64

	Role     MediaPlacementRole
	Position int

	Captioni string
	AltText  string

	CreatedAt time.Time
	UpdatedAt time.Time
}
