package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"flowerpress/internal/database"
	"flowerpress/internal/service"
	"flowerpress/internal/store/turso"
)

func testServer(t *testing.T) *Server {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"flowerpress.db",
	)

	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	users := turso.NewUserRepository(db)
	sessions := turso.NewSessionRepository(db)
	projects := turso.NewProjectRepository(db)

	return NewServer(
		service.NewUserService(users),
		service.NewSessionService(sessions, users, 24*time.Hour),
		service.NewProjectService(projects),
		false, // SecureCookie?
	)
}

func loginTestUser(t *testing.T, server *Server) *http.Cookie { // Refactor properly into tests later
	t.Helper()

	registerBody := `{"username":"flower","password":"newgarden"}`
	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(registerBody),
	)

	registerResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		registerResponse,
		registerRequest,
	)

	if registerResponse.Code != http.StatusCreated {
		t.Fatalf(
			"register test user: expected %d, got %d: %s",
			http.StatusCreated,
			registerResponse.Code,
			registerResponse.Body.String(),
		)
	}

	loginBody := `{"username":"flower","password":"newgarden"}`
	loginRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(loginBody),
	)

	loginResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		loginResponse,
		loginRequest,
	)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf(
			"login test user: expected %d, got %d: %s",
			http.StatusOK,
			loginResponse.Code,
			loginResponse.Body.String(),
		)
	}

	result := loginResponse.Result()
	defer result.Body.Close()

	cookies := result.Cookies() // Cut this out later, otherwise rename loginTestUserCookie

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 session cookie, got %d",
			len(cookies),
		)
	}

	return cookies[0]
}
