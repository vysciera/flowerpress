package httpapi

import (
	"net/http"
	"net/http/httptest"
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
