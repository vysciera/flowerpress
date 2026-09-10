package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"flowerpress/internal/domain"
	"flowerpress/internal/service"
)

const (
	maxMediaUploadSize  int64 = 100 << 20                      // 100 MiB
	maxMediaRequestSize int64 = maxMediaUploadSize + (1 << 20) // room for multipart headers + 100 MiB file
)

type mediaAssetResponse struct {
	ID int64 `json:"id"`

	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	SHA256       string `json:"sha256"`

	Width  *int `json:"width"`
	Height *int `json:"height"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type placeMediaRequest struct {
	AssetID  int64                     `json:"asset_id"`
	Role     domain.MediaPlacementRole `json:"role"`
	Position int                       `json:"position"`
	Caption  string                    `json:"caption"`
	AltText  string                    `json:"alt_text"`
}

type updateMediaPlacementRequest struct {
	Role     domain.MediaPlacementRole `json:"role"`
	Position int                       `json:"position"`
	Caption  string                    `json:"caption"`
	AltText  string                    `json:"alt_text"`
}

type mediaPlacementResponse struct {
	ID        int64                     `json:"id"`
	AssetID   int64                     `json:"asset_id"`
	ProjectID int64                     `json:"project_id"`
	Role      domain.MediaPlacementRole `json:"role"`
	Position  int                       `json:"position"`
	Caption   string                    `json:"caption"`
	AltText   string                    `json:"alt_text"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

type reorderProjectMediaRequest struct {
	Role         domain.MediaPlacementRole `json:"role"`
	PlacementIDs []int64                   `json:"placement_ids"`
}

type publicProjectMediaResponse struct {
	Placement mediaPlacementResponse   `json:"placement"`
	Asset     publicMediaAssetResponse `json:"asset"`
}

type publicMediaAssetResponse struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Width        *int   `json:"width"`
	Height       *int   `json:"height"`
}

func (s *Server) handleUpdateMediaPlacement(w http.ResponseWriter, r *http.Request) {
	placementID, err := mediaPlacementIDFromRequest(r)
	if err != nil || placementID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media placement id",
			},
		)
		return
	}

	var request updateMediaPlacementRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	placement, err := s.media.UpdatePlacement(
		r.Context(),
		placementID,
		request.Role,
		request.Position,
		request.Caption,
		request.AltText,
	)

	switch {
	case errors.Is(err, domain.ErrMediaPlacementNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media placement not found",
			},
		)
		return

	case errors.Is(err, service.ErrInvalidMediaPlacementRole),
		errors.Is(err, service.ErrInvalidMediaPosition):

		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
		return

	case errors.Is(err, service.ErrProjectThumbnailExists):
		writeJSON(
			w, http.StatusConflict,
			map[string]string{
				"error": err.Error(),
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w, http.StatusOK,
		mediaPlacementToResponse(placement),
	)
}

func (s *Server) handleDeleteMediaPlacement(w http.ResponseWriter, r *http.Request) {
	placementID, err := mediaPlacementIDFromRequest(r)
	if err != nil || placementID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media placement id",
			},
		)
		return
	}

	err = s.media.RemovePlacement(r.Context(), placementID)
	switch {
	case errors.Is(err, domain.ErrMediaPlacementNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media placement not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePlaceMedia(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	var request placeMediaRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	if request.AssetID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media asset id",
			},
		)
		return
	}

	placement, err := s.media.PlaceAsset(
		r.Context(),
		projectID,
		request.AssetID,
		request.Role,
		request.Position,
		request.Caption,
		request.AltText,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case errors.Is(err, domain.ErrMediaAssetNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media asset not found",
			},
		)
		return

	case errors.Is(err, service.ErrInvalidMediaPlacementRole),
		errors.Is(err, service.ErrInvalidMediaPosition):

		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "media asset not found",
			},
		)
		return

	case errors.Is(err, service.ErrProjectThumbnailExists):
		writeJSON(
			w, http.StatusConflict,
			map[string]string{
				"error": err.Error(),
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w, http.StatusCreated,
		mediaPlacementToResponse(placement),
	)
}

func (s *Server) handleGetMedia(w http.ResponseWriter, r *http.Request) {
	mediaID, err := mediaIDFromRequest(r)
	if err != nil || mediaID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media id",
			},
		)
		return
	}

	asset, err := s.media.AssetByID(r.Context(), mediaID)
	switch {
	case errors.Is(err, domain.ErrMediaAssetNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media asset not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w, http.StatusOK,
		mediaAssetToResponse(asset),
	)
}

func (s *Server) handleListMedia(w http.ResponseWriter, r *http.Request) {
	assets, err := s.media.ListAssets(r.Context())
	if err != nil {
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	// Empty responses return [], not nil
	response := make([]mediaAssetResponse, 0, len(assets))
	for _, asset := range assets {
		response = append(response, mediaAssetToResponse(asset))
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleMediaContent(w http.ResponseWriter, r *http.Request) {
	mediaID, err := mediaIDFromRequest(r)
	if err != nil || mediaID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media id",
			},
		)
		return
	}

	asset, content, err := s.media.OpenAssetContent(r.Context(), mediaID)
	switch {
	case errors.Is(err, domain.ErrMediaAssetNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media asset not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	defer content.Close()

	w.Header().Set("Content-Type", asset.MIMEType)

	// Incoming number of bytes
	w.Header().Set("Content-Length", strconv.FormatInt(
		asset.SizeBytes,
		10,
	))

	// Don't MIMESniff me
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set(
		"Content-Disposition",
		publicMediaDisposition(asset),
	)

	w.WriteHeader(http.StatusOK)

	// Headers may have already been sent
	// No real useful JSON error response to write (atm)
	if _, err := io.Copy(w, content); err != nil {
		return
	}
}

func (s *Server) handleListProjectMedia(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	placements, err := s.media.ListProjectMedia(r.Context(), projectID)
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	response := make([]mediaPlacementResponse, 0, len(placements))
	for _, placement := range placements {
		response = append(response, mediaPlacementToResponse(placement))
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleUploadMedia(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMediaRequestSize)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var maxErr *http.MaxBytesError

		if errors.As(err, &maxErr) {
			writeJSON(
				w, http.StatusRequestEntityTooLarge,
				map[string]string{
					"error": "media upload too large",
				},
			)
			return
		}

		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid multipart request",
			},
		)
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "file is required",
			},
		)
		return
	}
	defer file.Close()

	if header.Size > maxMediaUploadSize {
		writeJSON(
			w, http.StatusRequestEntityTooLarge,
			map[string]string{
				"error": "media upload too large",
			},
		)
		return
	}

	// Don't inherently trust MIME type supplied by client
	// Read first 512 bytes, let net/http inspect
	// Rewind multipart file before passing to MediaService
	headerBytes := make([]byte, 512)

	n, readErr := file.Read(headerBytes)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "could not inspect media file",
			},
		)
		return
	}

	mimeType := http.DetectContentType(headerBytes[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	asset, err := s.media.UploadAsset(
		r.Context(),
		header.Filename,
		mimeType,
		file,
		nil,
		nil,
	)

	switch {
	case errors.Is(err, service.ErrMediaOriginalNameRequired),
		errors.Is(err, service.ErrMediaMIMETypeRequired),
		errors.Is(err, service.ErrInvalidMediaDimensions):

		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w, http.StatusCreated,
		mediaAssetToResponse(asset),
	)
}

func (s *Server) handleReorderProjectMedia(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	var request reorderProjectMediaRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err = s.media.ReorderPlacements(
		r.Context(),
		projectID,
		request.Role,
		request.PlacementIDs,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case errors.Is(err, service.ErrInvalidMediaPlacementRole),
		errors.Is(err, service.ErrInvalidMediaOrder):

		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePublicProjectMedia(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	project, err := s.projects.ByPublicSlug(r.Context(), slug)
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	placements, err := s.media.ListProjectMedia(r.Context(), project.ID)
	if err != nil {
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	response := make([]publicProjectMediaResponse, 0, len(placements))
	for _, placement := range placements {
		asset, err := s.media.AssetByID(r.Context(), placement.AssetID)
		if err != nil {
			writeJSON(
				w, http.StatusInternalServerError,
				map[string]string{
					"error": "internal server error",
				},
			)
			return
		}

		response = append(
			response,
			publicProjectMediaResponse{
				Placement: mediaPlacementToResponse(placement),
				Asset:     publicMediaAssetToResponse(asset), // NOTE: Single JOIN query later
			},
		)
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handlePublicMediaContent(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	assetID, err := publicMediaAssetIDFromRequest(r)
	if err != nil || assetID <= 0 {
		writeJSON(
			w, http.StatusBadRequest,
			map[string]string{
				"error": "invalid media asset id",
			},
		)
		return
	}

	project, err := s.projects.ByPublicSlug(r.Context(), slug)
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	asset, content, err := s.media.OpenProjectAssetContent(
		r.Context(),
		project.ID,
		assetID,
	)

	switch {
	case errors.Is(err, service.ErrMediaAssetNotPlaced),
		errors.Is(err, domain.ErrMediaAssetNotFound),
		errors.Is(err, domain.ErrProjectNotFound):

		writeJSON(
			w, http.StatusNotFound,
			map[string]string{
				"error": "media not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	defer content.Close()

	w.Header().Set("Content-Type", asset.MIMEType)
	w.Header().Set("Content-Length", strconv.FormatInt(
		asset.SizeBytes, 10,
	))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set(
		"Content-Disposition",
		publicMediaDisposition(asset),
	)

	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, content)
}

func mediaAssetToResponse(asset *domain.MediaAsset) mediaAssetResponse {
	return mediaAssetResponse{
		ID:           asset.ID,
		OriginalName: asset.OriginalName,
		MIMEType:     asset.MIMEType,
		SizeBytes:    asset.SizeBytes,
		SHA256:       asset.SHA256,
		Width:        asset.Width,
		Height:       asset.Height,
		CreatedAt:    asset.CreatedAt,
		UpdatedAt:    asset.UpdatedAt,
	}
}

func publicMediaAssetToResponse(asset *domain.MediaAsset) publicMediaAssetResponse {
	return publicMediaAssetResponse{
		ID:           asset.ID,
		OriginalName: asset.OriginalName,
		MIMEType:     asset.MIMEType,
		SizeBytes:    asset.SizeBytes,
		Width:        asset.Width,
		Height:       asset.Height,
	}
}

func mediaPlacementToResponse(placement *domain.MediaPlacement) mediaPlacementResponse {
	return mediaPlacementResponse{
		ID:        placement.ID,
		AssetID:   placement.AssetID,
		ProjectID: placement.ProjectID,
		Role:      placement.Role,
		Position:  placement.Position,
		Caption:   placement.Caption,
		AltText:   placement.AltText,
		CreatedAt: placement.CreatedAt,
		UpdatedAt: placement.UpdatedAt,
	}
}

func mediaIDFromRequest(r *http.Request) (int64, error) {
	return strconv.ParseInt(
		chi.URLParam(r, "id"),
		10,
		64,
	)
}

func mediaPlacementIDFromRequest(r *http.Request) (int64, error) {
	return strconv.ParseInt(
		chi.URLParam(r, "id"),
		10, 64,
	)
}

func publicMediaAssetIDFromRequest(r *http.Request) (int64, error) {
	return strconv.ParseInt(
		chi.URLParam(r, "assetID"),
		10, 64,
	)
}

func publicMediaDisposition(asset *domain.MediaAsset) string {
	disposition := "attachment"

	switch {
	case strings.HasPrefix(asset.MIMEType, "audio/"):
		disposition = "inline"

	case strings.HasPrefix(asset.MIMEType, "image/") && asset.MIMEType != "image/svg+xml":
		disposition = "inline"

	case asset.MIMEType == "application/pdf":
		disposition = "inline"
	}

	return mime.FormatMediaType(
		disposition,
		map[string]string{
			"filename": asset.OriginalName,
		},
	)
}
