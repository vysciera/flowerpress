package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"flowerpress/internal/domain"
)

func multipartUpload(t *testing.T, filename, contentType string, content []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)

	header.Set(
		"Content-Disposition",
		`form-data; name="file"; filename="`+filename+`"`,
	)

	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestUploadMediaRequiresAuthentication(t *testing.T) {
	server := testServer(t)

	body, contentType := multipartUpload(
		t,
		"flower.txt",
		"text/plain",
		[]byte("flowerpress"),
	)

	request := httptest.NewRequest(http.MethodPost, "/api/media", body)
	request.Header.Set("Content-Type", contentType)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestUploadMedia(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	content := []byte("flowerpress media")

	body, contentType := multipartUpload(
		t,
		"flower.txt",
		"text/plain",
		content,
	)

	request := httptest.NewRequest(http.MethodPost, "/api/media", body)
	request.Header.Set("Content-Type", contentType)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var asset mediaAssetResponse

	if err := json.NewDecoder(response.Body).Decode(&asset); err != nil {
		t.Fatalf("decode media asset: %v", err)
	}

	if asset.ID == 0 {
		t.Fatal("expected media asset ID")
	}

	if asset.OriginalName != "flower.txt" {
		t.Fatalf(
			"expected filename %q, got %q",
			"flower.txt",
			asset.OriginalName,
		)
	}

	if asset.SizeBytes != int64(len(content)) {
		t.Fatalf(
			"expected size %d, got %d",
			len(content),
			asset.SizeBytes,
		)
	}

	if asset.SHA256 == "" {
		t.Fatal("expected SHA256")
	}

	if asset.MIMEType != "text/plain; charset=utf-8" {
		t.Fatalf(
			"expected detected MIME type %q, got %q",
			"text/plain; charset=utf-8",
			asset.MIMEType,
		)
	}
}

func TestUploadMediaRequiresFile(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/media",
		body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestUploadMediaDeduplicates(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	content := []byte("identical flowerpress content")

	upload := func(filename string) mediaAssetResponse {
		t.Helper()

		body, contentType := multipartUpload(
			t,
			filename,
			"text/plain",
			content,
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/api/media",
			body,
		)

		request.Header.Set("Content-Type", contentType)

		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)

		if response.Code != http.StatusCreated {
			t.Fatalf(
				"upload %q: expected status %d, got %d: %s",
				filename,
				http.StatusCreated,
				response.Code,
				response.Body.String(),
			)
		}

		var asset mediaAssetResponse

		if err := json.NewDecoder(response.Body).Decode(&asset); err != nil {
			t.Fatalf("decode media asset: %v", err)
		}

		return asset
	}

	first := upload("first.txt")
	second := upload("second.txt")

	if first.ID != second.ID {
		t.Fatalf(
			"expected asset ID %d, got %d",
			first.ID,
			second.ID,
		)
	}

	if second.OriginalName != "first.txt" {
		t.Fatalf(
			"expected original asset metadata to be reused, got %q",
			second.OriginalName,
		)
	}
}

func TestListMediaRequiresAuthentication(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media",
		nil,
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestListMediaEmpty(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	var assets []mediaAssetResponse

	if err := json.NewDecoder(response.Body).Decode(&assets); err != nil {
		t.Fatalf("decode media list: %v", err)
	}

	if len(assets) != 0 {
		t.Fatalf(
			"expected empty media library, got %d assets",
			len(assets),
		)
	}
}

func TestListMedia(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	upload := func(filename string, content []byte) mediaAssetResponse {
		t.Helper()

		body, contentType := multipartUpload(
			t,
			filename,
			"text/plain",
			content,
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/api/media",
			body,
		)

		request.Header.Set(
			"Content-Type",
			contentType,
		)

		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)

		if response.Code != http.StatusCreated {
			t.Fatalf(
				"upload %q: expected %d, got %d: %s",
				filename,
				http.StatusCreated,
				response.Code,
				response.Body.String(),
			)
		}

		var asset mediaAssetResponse

		if err := json.NewDecoder(response.Body).Decode(&asset); err != nil {
			t.Fatalf("decode uploaded asset: %v", err)
		}

		return asset
	}

	first := upload(
		"first.txt",
		[]byte("first flower"),
	)

	second := upload(
		"second.txt",
		[]byte("second flower"),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	var assets []mediaAssetResponse

	if err := json.NewDecoder(response.Body).Decode(&assets); err != nil {
		t.Fatalf("decode media list: %v", err)
	}

	if len(assets) != 2 {
		t.Fatalf("expected 2 media assets, got %d", len(assets))
	}

	ids := make(map[int64]bool)
	for _, asset := range assets {
		ids[asset.ID] = true
	}

	if !ids[first.ID] {
		t.Fatalf("expected media asset ID %d", first.ID)
	}

	if !ids[second.ID] {
		t.Fatalf("expected media asset ID %d", second.ID)
	}
}

func TestGetMediaRequiresAuthentication(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media/1",
		nil,
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestGetMedia(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	content := []byte(
		"flowerpress media",
	)

	body, contentType := multipartUpload(
		t,
		"flower.txt",
		"text/plain",
		content,
	)

	uploadRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/media",
		body,
	)

	uploadRequest.Header.Set("Content-Type", contentType)

	uploadRequest.AddCookie(cookie)
	uploadResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(uploadResponse, uploadRequest)

	if uploadResponse.Code != http.StatusCreated {
		t.Fatalf(
			"upload media: expected %d, got %d: %s",
			http.StatusCreated,
			uploadResponse.Code,
			uploadResponse.Body.String(),
		)
	}

	var uploaded mediaAssetResponse

	if err := json.NewDecoder(uploadResponse.Body).Decode(&uploaded); err != nil {
		t.Fatalf("decode uploaded media: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/api/media/%d",
			uploaded.ID,
		),
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	var found mediaAssetResponse

	if err := json.NewDecoder(response.Body).Decode(&found); err != nil {
		t.Fatalf("decode media asset: %v", err)
	}

	if found.ID != uploaded.ID {
		t.Fatalf(
			"expected media ID %d, got %d",
			uploaded.ID,
			found.ID,
		)
	}

	if found.OriginalName != "flower.txt" {
		t.Fatalf(
			"expected original name %q, got %q",
			"flower.txt",
			found.OriginalName,
		)
	}
}

func TestGetMediaNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media/999",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestGetMediaRejectsInvalidID(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	for _, path := range []string{
		"/api/media/nope",
		"/api/media/0",
		"/api/media/-1",
	} {
		request := httptest.NewRequest(
			http.MethodGet,
			path,
			nil,
		)

		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf(
				"%s: expected status %d, got %d",
				path,
				http.StatusBadRequest,
				response.Code,
			)
		}
	}
}

func TestMediaContent(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	expected := []byte(
		"flowerpress stored media",
	)

	body, contentType := multipartUpload(
		t,
		"flower.txt",
		"text/plain",
		expected,
	)

	uploadRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/media",
		body,
	)

	uploadRequest.Header.Set("Content-Type", contentType)

	uploadRequest.AddCookie(cookie)
	uploadResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(uploadResponse, uploadRequest)

	if uploadResponse.Code != http.StatusCreated {
		t.Fatalf(
			"upload media: expected %d, got %d: %s",
			http.StatusCreated,
			uploadResponse.Code,
			uploadResponse.Body.String(),
		)
	}

	var asset mediaAssetResponse

	if err := json.NewDecoder(
		uploadResponse.Body,
	).Decode(&asset); err != nil {
		t.Fatalf(
			"decode uploaded media: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/api/media/%d/content",
			asset.ID,
		),
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !bytes.Equal(response.Body.Bytes(), expected) {
		t.Fatalf(
			"expected content %q, got %q",
			expected,
			response.Body.Bytes(),
		)
	}

	if response.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf(
			"unexpected Content-Type %q",
			response.Header().Get("Content-Type"),
		)
	}

	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected X-Content-Type-Options nosniff")
	}
}

func TestMediaContentRequiresAuthentication(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media/1/content",
		nil,
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestMediaContentNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/media/999/content",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestPlaceMedia(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	// Create project.
	projectRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/projects",
		jsonBody(t, map[string]string{
			"title": "Test Project",
		}),
	)

	projectRequest.Header.Set("Content-Type", "application/json")
	projectRequest.AddCookie(cookie)
	projectResponseRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(projectResponseRecorder, projectRequest)

	if projectResponseRecorder.Code != http.StatusCreated {
		t.Fatalf(
			"create project: expected %d, got %d: %s",
			http.StatusCreated,
			projectResponseRecorder.Code,
			projectResponseRecorder.Body.String(),
		)
	}

	var project projectResponse

	if err := json.NewDecoder(projectResponseRecorder.Body).Decode(&project); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	// Upload asset.
	body, contentType := multipartUpload(
		t,
		"flower.txt",
		"text/plain",
		[]byte("flowerpress"),
	)

	uploadRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/media",
		body,
	)

	uploadRequest.Header.Set("Content-Type", contentType)
	uploadRequest.AddCookie(cookie)
	uploadResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(uploadResponse, uploadRequest)

	if uploadResponse.Code != http.StatusCreated {
		t.Fatalf(
			"upload media: expected %d, got %d: %s",
			http.StatusCreated,
			uploadResponse.Code,
			uploadResponse.Body.String(),
		)
	}

	var asset mediaAssetResponse

	if err := json.NewDecoder(uploadResponse.Body).Decode(&asset); err != nil {
		t.Fatalf("decode asset: %v", err)
	}

	// Place asset in project.
	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/media",
			project.ID,
		),
		jsonBody(t, placeMediaRequest{
			AssetID:  asset.ID,
			Role:     domain.MediaPlacementContent,
			Position: 0,
			Caption:  "A flower",
			AltText:  "Flower",
		}),
	)

	request.Header.Set("Content-Type", "application/json")

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var placement mediaPlacementResponse

	if err := json.NewDecoder(response.Body).Decode(&placement); err != nil {
		t.Fatalf("decode placement: %v", err)
	}

	if placement.AssetID != asset.ID {
		t.Fatalf(
			"expected asset ID %d, got %d",
			asset.ID,
			placement.AssetID,
		)
	}

	if placement.ProjectID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			placement.ProjectID,
		)
	}

	if placement.Role != domain.MediaPlacementContent {
		t.Fatalf("expected content role, got %q", placement.Role)
	}
}
