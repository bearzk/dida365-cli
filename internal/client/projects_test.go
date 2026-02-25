package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/bearzk/dida365-cli/internal/models"
)

func TestListProjects(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		projects := []models.Project{
			{
				ID:        "proj1",
				Name:      "Personal",
				Color:     "#FF0000",
				SortOrder: 1,
				Closed:    false,
				Kind:      "TASK",
			},
			{
				ID:        "proj2",
				Name:      "Work",
				Color:     "#00FF00",
				SortOrder: 2,
				Closed:    false,
				Kind:      "TASK",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project" {
				t.Errorf("path: got %s, want /open/v1/project", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(projects)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 projects, got %d", len(result))
		}

		if result[0].ID != "proj1" {
			t.Errorf("project 0 ID: got %s, want proj1", result[0].ID)
		}
		if result[1].Name != "Work" {
			t.Errorf("project 1 Name: got %s, want Work", result[1].Name)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]models.Project{})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected empty list, got %d projects", len(result))
		}
	})

	t.Run("unauthorized error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "invalid",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		_, err := client.ListProjects()
		if err == nil {
			t.Fatal("expected error for unauthorized, got nil")
		}

		if !contains(err.Error(), "access token") {
			t.Errorf("error message should mention access token, got: %s", err.Error())
		}
	})
}

func TestGetProject(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		project := models.Project{
			ID:        "proj123",
			Name:      "Test Project",
			Color:     "#0000FF",
			SortOrder: 5,
			Closed:    false,
			Kind:      "TASK",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project/proj123" {
				t.Errorf("path: got %s, want /open/v1/project/proj123", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(project)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.GetProject("proj123")
		if err != nil {
			t.Fatalf("GetProject failed: %v", err)
		}

		if result.ID != "proj123" {
			t.Errorf("ID: got %s, want proj123", result.ID)
		}
		if result.Name != "Test Project" {
			t.Errorf("Name: got %s, want Test Project", result.Name)
		}
	})

	t.Run("not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		_, err := client.GetProject("nonexistent")
		if err == nil {
			t.Fatal("expected error for not found, got nil")
		}

		if !contains(err.Error(), "not found") {
			t.Errorf("error should mention not found, got: %s", err.Error())
		}
	})
}

func TestGetProjectData(t *testing.T) {
	t.Run("returns raw json bytes", func(t *testing.T) {
		rawResponse := `{"tasks":[{"id":"t1","title":"Test"}],"columns":[{"id":"c1","name":"To Do"}]}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project/proj123/data" {
				t.Errorf("path: got %s, want /open/v1/project/proj123/data", r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test" {
				t.Errorf("Authorization: got %q, want %q", got, "Bearer test")
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, rawResponse)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		c := NewClient(cfg)
		result, err := c.GetProjectData("proj123")
		if err != nil {
			t.Fatalf("GetProjectData failed: %v", err)
		}

		if string(result) != rawResponse {
			t.Errorf("raw bytes: got %s, want %s", string(result), rawResponse)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"not found"}`)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		c := NewClient(cfg)
		_, err := c.GetProjectData("missing")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
