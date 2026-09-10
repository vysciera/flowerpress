package httpapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func multipartUpload(t *testing.T, filename, contentType string, content []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)

	header.Set(
		"Content-Disposition",
		`form-data; name="file"; filename="` + filename + `"`,
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
