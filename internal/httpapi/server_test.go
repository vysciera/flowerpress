package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	server := NewServer(nil, nil, nil, false)

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

func TestLogout(t *testing.T) {
	server := testServer(t)

	register := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	server.Handler().ServeHTTP( // Normalize test structures later cause this is driving me insane
		httptest.NewRecorder(),
		register,
	)

	login := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	loginResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		loginResponse,
		login,
	)

	result := loginResponse.Result()
	defer result.Body.Close()

	cookies := result.Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected session cookie, got %d cookies",
			len(cookies),
		)
	}

	// coocie.
	sessionCookie := cookies[0]

	logout := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/logout",
		nil,
	)

	logout.AddCookie(sessionCookie)

	logoutResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		logoutResponse,
		logout,
	)

	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			logoutResponse.Code,
		)
	}
}

func TestLogoutClearsSessionCookie(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/logout",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "some-token",
	})

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	result := response.Result()
	defer result.Body.Close()

	cookies := result.Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 cookie, got %d",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if cookie.Name != sessionCookieName {
		t.Fatalf(
			"expected cookie %q, got %q",
			sessionCookieName,
			cookie.Name,
		)
	}

	if cookie.MaxAge >= 0 {
		t.Fatalf(
			"expected cookie to be expired, got MaxAge %d",
			cookie.MaxAge,
		)
	}
}

func TestLogoutWithoutSession(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/logout",
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"unexpected status %d, got %d",
			http.StatusNoContent,
			response.Code,
		)
	}
}

func TestLogin(t *testing.T) {
	server := testServer(t)

	registerBody := strings.NewReader(`
		{
			"username": "flower",
			"password": "newgarden"
		}
	`)

	registerRequest := httptest.NewRequest(http.MethodPost, "/api/auth/register", registerBody)
	registerResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register returned %d, %s", registerResponse.Code, registerResponse.Body.String())
	}

	loginBody := strings.NewReader(`
		{
			"username": "flower",
			"password": "newgarden"
		}
	`)

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}
}

func TestLoginSetsSessionCookie(t *testing.T) {
	server := testServer(t)

	register := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	server.Handler().ServeHTTP(
		httptest.NewRecorder(),
		register,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	// Just one cookie bro, please bro.
	cookie := cookies[0]

	if cookie.Name != sessionCookieName {
		t.Fatalf("expected cookie %q, got %q", sessionCookieName, cookie.Name)
	}

	if cookie.Value == "" {
		t.Fatal("expected session cookie value")
	}

	if !cookie.HttpOnly {
		t.Fatal("expected session cookie to be HttpOnly")
	}

	if cookie.Path != "/" {
		t.Fatalf("expected cookie path %q, got %q", "/", cookie.Path)
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite Lax, got %v", cookie.SameSite)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	server := testServer(t)

	register := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	server.Handler().ServeHTTP(
		httptest.NewRecorder(),
		register,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(
			`{"username":"flower","password":"badpwd-garden"}`,
		),
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expect status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(
			`{"username":"nobody","password":"nopassword"}`,
		),
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
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

func TestMeRequiresAuthentication(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"xpected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestMeRejectsInvalidSession(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/auth/me",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "invalid-session-token",
	})

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	server := testServer(t)

	register := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	registerResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		registerResponse,
		register,
	)

	if registerResponse.Code != http.StatusCreated {
		t.Fatalf(
			"register returned %d: %s",
			registerResponse.Code,
			registerResponse.Body.String(),
		)
	}

	login := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(
			`{"username":"flower","password":"newgarden"}`,
		),
	)

	loginResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		loginResponse,
		login,
	)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf(
			"login returned %d: %s",
			loginResponse.Code,
			loginResponse.Body.String(),
		)
	}

	result := loginResponse.Result()
	defer result.Body.Close()

	cookies := result.Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected session cookie, got %d cookies",
			len(cookies),
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/auth/me",
		nil,
	)

	request.AddCookie(cookies[0])

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	expected := `"username":"flower"`

	if !strings.Contains(
		response.Body.String(),
		expected,
	) {
		t.Fatalf(
			"expected response to contain %q, got %s",
			expected,
			response.Body.String(),
		)
	}
}

func TestRegisterRejectsUnknownFields(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(`
			{
				"username": "flower",
				"password": "newgarden",
				"admin": true
			}	
		`),
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestRegisterRejectsMultipleJSONObjects(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(`
			{"username":"flower","password":"newgarden"}
			{"username":"garden","password":"fakegarden"}
		`),
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestRegisterRejectsOversizedBody(t *testing.T) {
	server := testServer(t)

	body := `{"username":"` +
		strings.Repeat("a", maxRequestBodySize+1) +
		`","password":"newgarden"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(body),
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusRequestEntityTooLarge,
			response.Code,
		)
	}
}

func TestLoginSecureCookie(t *testing.T) {
	server := testServer(t)
	server.secureCookies = true

	cookie := loginTestUser(t, server)

	if !cookie.Secure {
		t.Fatal("expected session cookie to be secure")
	}
}

func TestLoginDevelopmentCookieNotSecure(t *testing.T) {
	server := testServer(t)
	server.secureCookies = false

	cookie := loginTestUser(t, server)

	if cookie.Secure {
		t.Fatal("expected development session cookie to not be secure")
	}
}
