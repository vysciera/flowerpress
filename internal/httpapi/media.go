package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strconv"
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

	writeJSON(
		w, http.StatusOK,
		response,
	)
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
	w.WriteHeader(http.StatusOK)

	// Headers may have already been sent
	// No real useful JSON error response to write (atm)
	if _, err := io.Copy(w, content); err != nil {
		return
	}
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

func mediaIDFromRequest(r *http.Request) (int64, error) {
	return strconv.ParseInt(
		chi.URLParam(r, "id"),
		10,
		64,
	)
}
