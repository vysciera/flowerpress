package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	userRepository := turso.NewUserRepository(db)
	sessionRepository := turso.NewSessionRepository(db)
	projectRepository := turso.NewProjectRepository(db)
	users := service.NewUserService(userRepository)

	sessions := service.NewSessionService(
		sessionRepository,
		userRepository,
		24*time.Hour,
	)

	projects := service.NewProjectService(projectRepository)

	return NewServer(users, sessions, projects, false)
}

func jsonBody(t *testing.T, value any) io.Reader {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON body: %v", err)
	}

	return bytes.NewReader(data)
}

func loginTestUser(t *testing.T, server *Server) *http.Cookie {
	t.Helper()

	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		jsonBody(
			t,
			map[string]string{
				"username": "flower",
				"password": "newgarden",
			},
		),
	)

	registerResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(registerResponse, registerRequest)

	if registerResponse.Code != http.StatusCreated {
		t.Fatalf(
			"register test user: expected %d, got %d: %s",
			http.StatusCreated,
			registerResponse.Code,
			registerResponse.Body.String(),
		)
	}

	loginRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		jsonBody(
			t,
			map[string]string{
				"username": "flower",
				"password": "newgarden",
			},
		),
	)

	loginResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf(
			"login test user: expected %d, got %d: %s",
			http.StatusOK,
			loginResponse.Code,
			loginResponse.Body.String(),
		)
	}

	for _, cookie := range loginResponse.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			return cookie
		}
	}

	t.Fatal("login response did not contain session cookie")

	return nil
}

func createProjectResponse(t *testing.T, server *Server, cookie *http.Cookie, title string) projectResponse {
	t.Helper()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/projects",
		jsonBody(
			t,
			createProjectRequest{
				Title: title,
			},
		),
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"create test project: expected %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var project projectResponse

	if err := json.NewDecoder(response.Body).Decode(&project); err != nil {
		t.Fatalf("decode created project: %v", err)
	}

	return project
}
