package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	server := NewServer(nil, nil)

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	expected := `{"status":"ok"}`

	if response.Body.String() != expected {
		t.Fatalf("expected body %q, got %q", expected, response.Body.String())
	}
}

func TestRegister(t *testing.T) {
	server := testServer(t)

	body := strings.NewReader(`
		{
			"username": "flower",
			"password": "newgarden"
		}
	`)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	server := testServer(t)

	register := func(username string) *httptest.ResponseRecorder {
		body := strings.NewReader(
			`{"username":"` +
				username +
				`","password":"garden123"}`,
		)

		request := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
		response := httptest.NewRecorder()

		server.Handler().ServeHTTP(response, request)

		return response
	}

	first := register("flower")
	if first.Code != http.StatusCreated {
		t.Fatalf("first registration returned %d", first.Code)
	}

	second := register("FLOWER")
	if second.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, second.Code)
	}
}

func TestRegisterInvalidJSON(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(`{"username":}`),
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
