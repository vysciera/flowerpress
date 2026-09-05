package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"flowerpress/internal/domain"
)

func TestCreateProjectRequiresAuthentication(t *testing.T) {
	server := testServer(t)
	payload := `{"title":"Flowerpress"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/projects",
		strings.NewReader(payload),
	)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	endpoint := "/api/projects"
	payload := `{"title": "Flower Press", "description": "Personal archive"}`

	request := httptest.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(payload),
	)

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

	if !strings.Contains(
		response.Body.String(),
		`"slug":"flower-press"`,
	) {
		t.Fatalf("expected generated slug, got %s", response.Body.String())
	}

	if !strings.Contains(
		response.Body.String(),
		`"status":"draft"`,
	) {
		t.Fatalf("expected draft status, got %s", response.Body.String())
	}
}

func TestDeleteProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(
			"/api/projects/%d",
			project.ID,
		),
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNoContent,
			response.Code,
			response.Body.String(),
		)
	}

	get := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/api/projects/%d",
			project.ID,
		),
		nil,
	)

	get.AddCookie(cookie)
	getResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(getResponse, get)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected deleted project to return %d, got %d",
			http.StatusNotFound,
			getResponse.Code,
		)
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/projects/999",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.Code,
		)
	}
}

func TestCreateProjectRequiresTitle(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	endpoint := "/api/projects"
	payload := `{"title":"   "}`

	request := httptest.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(payload),
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

func TestListProjects(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	endpoint := "/api/projects"

	for _, title := range []string{"First", "Second"} {
		request := httptest.NewRequest(
			http.MethodPost,
			endpoint,
			strings.NewReader(
				`{"title":"`+title+`"}`,
			),
		)

		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)

		if response.Code != http.StatusCreated {
			t.Fatalf(
				"create project %q: %d: %s",
				title,
				response.Code,
				response.Body.String(),
			)
		}
	}

	request := httptest.NewRequest(
		http.MethodGet,
		endpoint,
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

	if !strings.Contains(
		response.Body.String(),
		`"title":"First"`,
	) {
		t.Fatalf("expected First project: %s", response.Body.String())
	}

	if !strings.Contains(
		response.Body.String(),
		`"title":"Second"`,
	) {
		t.Fatalf("expected Second project: %s", response.Body.String())
	}
}

func TestListProjectsEmpty(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	endpoint := "/api/projects"

	request := httptest.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	got := strings.TrimSpace(response.Body.String())
	if got != "[]" {
		t.Fatalf("expected empty array, got %s", got)
	}
}

func TestGetProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	endpoint := "/api/projects"
	payload := `{"title":"Flowerpress","description":"archive"}`

	createRequest := httptest.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(payload),
	)

	createRequest.AddCookie(cookie)
	createResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create project: %d: %s", createResponse.Code, createResponse.Body.String())
	}

	var created projectResponse

	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created project: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/api/projects/%d",
			created.ID,
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

	var found projectResponse
	if err := json.NewDecoder(response.Body).Decode(&found); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, found.ID)
	}

	if found.Title != "Flowerpress" {
		t.Fatalf("expected title %q, got %q", "Flowerpress", found.Title)
	}
}

func TestGetProjectNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/projects/999",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestGetProjectRejectsInvalidID(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/projects/blahblahblah",
		nil,
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestUpdateProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/projects",
		strings.NewReader(
			`{"title":"Flowerpress","description":"Old"}`,
		),
	)

	createRequest.AddCookie(cookie)
	createResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(createResponse, createRequest)

	var created projectResponse

	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created project: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf(
			"/api/projects/%d",
			created.ID,
		),
		strings.NewReader(`
			{
				"title": "Flowerpress Archive",
				"description": "New description"
			}	
		`),
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var updated projectResponse

	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated project: %v", err)
	}

	if updated.Title != "Flowerpress Archive" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	if updated.Description != "New description" {
		t.Fatalf("expected updated description, got %q", updated.Description)
	}

	if updated.Slug != created.Slug {
		t.Fatalf("expected slug %q to remain stable, got %q", created.Slug, updated.Slug)
	}
}

func TestUpdateProjectNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/projects/999",
		strings.NewReader(`
			{"title":"Missing"}	
		`),
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestUpdateProjectRequiresTitle(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/projects",
		strings.NewReader(
			`{"title":"Flowerpress"}`,
		),
	)

	createRequest.AddCookie(cookie)
	createResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(createResponse, createRequest)

	var created projectResponse

	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf(
			"/api/projects/%d",
			created.ID,
		),
		strings.NewReader(
			`{"title":"   "}`,
		),
	)

	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

// !!Project lifecycle handling
func TestPublishProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/publish",
			project.ID,
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

	var updated projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	if updated.Status != domain.ProjectStatusPublished {
		t.Fatalf(
			"expected published status, got %q",
			updated.Status,
		)
	}

	if updated.PublishedAt == nil {
		t.Fatal("expected PublishedAt")
	}
}

func TestUnpublishProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	publish := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/publish",
			project.ID,
		),
		nil,
	)
	publish.AddCookie(cookie)

	publishResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		publishResponse,
		publish,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/unpublish",
			project.ID,
		),
		nil,
	)
	request.AddCookie(cookie)

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

	var updated projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	if updated.Status != domain.ProjectStatusDraft {
		t.Fatalf(
			"expected draft status, got %q",
			updated.Status,
		)
	}
}

func TestUnlistProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/unlist",
			project.ID,
		),
		nil,
	)
	request.AddCookie(cookie)

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

	var updated projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	if updated.Status != domain.ProjectStatusUnlisted {
		t.Fatalf(
			"expected unlisted status, got %q",
			updated.Status,
		)
	}
}

func TestArchiveProject(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/archive",
			project.ID,
		),
		nil,
	)
	request.AddCookie(cookie)

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

	var updated projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	if updated.Status != domain.ProjectStatusArchived {
		t.Fatalf(
			"expected archived status, got %q",
			updated.Status,
		)
	}
}

func TestPublishArchivedProjectReturnsConflict(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	archive := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/archive",
			project.ID,
		),
		nil,
	)
	archive.AddCookie(cookie)

	archiveResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		archiveResponse,
		archive,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/publish",
			project.ID,
		),
		nil,
	)
	request.AddCookie(cookie)

	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestPublicProjectPublished(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	publishRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/publish",
			project.ID,
		),
		nil,
	)

	publishRequest.AddCookie(cookie)
	publishResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(publishResponse, publishRequest)

	if publishResponse.Code != http.StatusOK {
		t.Fatalf(
			"publish project: expected %d, got %d: %s",
			http.StatusOK,
			publishResponse.Code,
			publishResponse.Body.String(),
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects/"+project.Slug,
		nil,
	)

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

	var found projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&found); err != nil {
		t.Fatalf("decode public project: %v", err)
	}

	if found.ID != project.ID {
		t.Fatalf(
			"expected project ID %d, got %d",
			project.ID,
			found.ID,
		)
	}
}

func TestPublicProjectDraftReturnsNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects/"+project.Slug,
		nil,
	)

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

func TestPublicProjectUnlisted(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	unlistRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/unlist",
			project.ID,
		),
		nil,
	)

	unlistRequest.AddCookie(cookie)
	unlistResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(unlistResponse, unlistRequest)

	if unlistResponse.Code != http.StatusOK {
		t.Fatalf(
			"unlist project: expected %d, got %d: %s",
			http.StatusOK,
			unlistResponse.Code,
			unlistResponse.Body.String(),
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects/"+project.Slug,
		nil,
	)

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
}

func TestPublicProjectArchivedReturnsNotFound(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	project := createProjectResponse(
		t,
		server,
		cookie,
		"Flowerpress",
	)

	archiveRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/archive",
			project.ID,
		),
		nil,
	)

	archiveRequest.AddCookie(cookie)
	archiveResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(archiveResponse, archiveRequest)

	if archiveResponse.Code != http.StatusOK {
		t.Fatalf(
			"archive project: expected %d, got %d: %s",
			http.StatusOK,
			archiveResponse.Code,
			archiveResponse.Body.String(),
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects/"+project.Slug,
		nil,
	)

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

func TestPublicProjectNotFound(t *testing.T) {
	server := testServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects/does-not-exist",
		nil,
	)

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.Code,
		)
	}
}

func TestPublicProjects(t *testing.T) {
	server := testServer(t)
	cookie := loginTestUser(t, server)

	published := createProjectResponse(
		t,
		server,
		cookie,
		"Published",
	)

	unlisted := createProjectResponse(
		t,
		server,
		cookie,
		"Unlisted",
	)

	draft := createProjectResponse(
		t,
		server,
		cookie,
		"Draft",
	)

	publishRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/publish",
			published.ID,
		),
		nil,
	)
	publishRequest.AddCookie(cookie)

	publishResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		publishResponse,
		publishRequest,
	)

	if publishResponse.Code != http.StatusOK {
		t.Fatalf(
			"publish project: %d: %s",
			publishResponse.Code,
			publishResponse.Body.String(),
		)
	}

	unlistRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/api/projects/%d/unlist",
			unlisted.ID,
		),
		nil,
	)
	unlistRequest.AddCookie(cookie)

	unlistResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		unlistResponse,
		unlistRequest,
	)

	if unlistResponse.Code != http.StatusOK {
		t.Fatalf(
			"unlist project: %d: %s",
			unlistResponse.Code,
			unlistResponse.Body.String(),
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/public/projects",
		nil,
	)

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

	var found []projectResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&found); err != nil {
		t.Fatalf(
			"decode public projects: %v",
			err,
		)
	}

	if len(found) != 1 {
		t.Fatalf(
			"expected 1 public project, got %d",
			len(found),
		)
	}

	if found[0].ID != published.ID {
		t.Fatalf(
			"expected published project ID %d, got %d",
			published.ID,
			found[0].ID,
		)
	}

	_ = draft
}
