// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package groups

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsers(t *testing.T) {
	mockUsers := []User{
		{
			ID:       "user-1",
			Name:     "Test User",
			Email:    "test@example.com",
			Role:     "user",
			IsActive: true,
			Groups:   []string{},
		},
		{
			ID:       "user-2",
			Name:     "Test Admin",
			Email:    "admin@example.com",
			Role:     "admin",
			IsActive: true,
			Groups:   []string{},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/groups/id/test-group-id/users" {
			t.Errorf("Expected path '/api/v1/groups/id/test-group-id/users', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer token, got %s", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockUsers)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	users, err := client.GetUsers("test-group-id")

	if err != nil {
		t.Fatalf("GetUsers returned error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].ID != "user-1" {
		t.Errorf("Expected first user ID 'user-1', got '%s'", users[0].ID)
	}

	if users[1].Email != "admin@example.com" {
		t.Errorf("Expected second user email 'admin@example.com', got '%s'", users[1].Email)
	}
}

func TestAddUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/groups/id/test-group-id/users/add" {
			t.Errorf("Expected path '/api/v1/groups/id/test-group-id/users/add', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var form UserIdsForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(form.UserIDs) != 2 {
			t.Errorf("Expected 2 user IDs, got %d", len(form.UserIDs))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&Group{ID: "test-group-id"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.AddUsers("test-group-id", []string{"user-1", "user-2"})

	if err != nil {
		t.Fatalf("AddUsers returned error: %v", err)
	}
}
