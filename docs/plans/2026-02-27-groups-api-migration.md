# Groups API Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate group user management from deprecated user_ids field to new /users, /users/add, and /users/remove endpoints.

**Architecture:** Add three new client methods (GetUsers, AddUsers, RemoveUsers) and orchestrate calls in resource CRUD operations to handle user membership separately from group metadata.

**Tech Stack:** Go, Terraform Plugin Framework, Open WebUI REST API

---

## Task 1: Add User type to client types

**Files:**
- Modify: `internal/provider/client/groups/types.go:16`

**Step 1: Add User struct after Group struct**

Add this code after the Group struct definition:

```go
type User struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Role            string   `json:"role"`
	StatusEmoji     *string  `json:"status_emoji"`
	StatusMessage   *string  `json:"status_message"`
	StatusExpiresAt *int64   `json:"status_expires_at"`
	Bio             *string  `json:"bio"`
	Groups          []string `json:"groups"`
	IsActive        bool     `json:"is_active"`
}

type UserIdsForm struct {
	UserIDs []string `json:"user_ids"`
}
```

**Step 2: Verify code compiles**

Run: `go build`
Expected: Successful compilation with no errors

**Step 3: Commit**

```bash
git add internal/provider/client/groups/types.go
git commit -m "feat(groups): add User and UserIdsForm types for new API"
```

---

## Task 2: Add GetUsers client method

**Files:**
- Modify: `internal/provider/client/groups/client.go:166`

**Step 1: Write test for GetUsers**

Create: `internal/provider/client/groups/client_test.go`

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/client/groups/... -v -run TestGetUsers`
Expected: FAIL with "client.GetUsers undefined"

**Step 3: Implement GetUsers method**

Add to `internal/provider/client/groups/client.go` after the List method:

```go
func (c *Client) GetUsers(id string) ([]User, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users", c.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/client/groups/... -v -run TestGetUsers`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/client/groups/client.go internal/provider/client/groups/client_test.go
git commit -m "feat(groups): add GetUsers client method"
```

---

## Task 3: Add AddUsers client method

**Files:**
- Modify: `internal/provider/client/groups/client.go`
- Modify: `internal/provider/client/groups/client_test.go`

**Step 1: Write test for AddUsers**

Add to `internal/provider/client/groups/client_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/client/groups/... -v -run TestAddUsers`
Expected: FAIL with "client.AddUsers undefined"

**Step 3: Implement AddUsers method**

Add to `internal/provider/client/groups/client.go` after GetUsers:

```go
func (c *Client) AddUsers(id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil // No-op for empty list
	}

	form := UserIdsForm{UserIDs: userIDs}
	body, err := json.Marshal(form)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users/add", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/client/groups/... -v -run TestAddUsers`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/client/groups/client.go internal/provider/client/groups/client_test.go
git commit -m "feat(groups): add AddUsers client method"
```

---

## Task 4: Add RemoveUsers client method

**Files:**
- Modify: `internal/provider/client/groups/client.go`
- Modify: `internal/provider/client/groups/client_test.go`

**Step 1: Write test for RemoveUsers**

Add to `internal/provider/client/groups/client_test.go`:

```go
func TestRemoveUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/groups/id/test-group-id/users/remove" {
			t.Errorf("Expected path '/api/v1/groups/id/test-group-id/users/remove', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var form UserIdsForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(form.UserIDs) != 1 {
			t.Errorf("Expected 1 user ID, got %d", len(form.UserIDs))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&Group{ID: "test-group-id"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.RemoveUsers("test-group-id", []string{"user-1"})

	if err != nil {
		t.Fatalf("RemoveUsers returned error: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/client/groups/... -v -run TestRemoveUsers`
Expected: FAIL with "client.RemoveUsers undefined"

**Step 3: Implement RemoveUsers method**

Add to `internal/provider/client/groups/client.go` after AddUsers:

```go
func (c *Client) RemoveUsers(id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil // No-op for empty list
	}

	form := UserIdsForm{UserIDs: userIDs}
	body, err := json.Marshal(form)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users/remove", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/client/groups/... -v -run TestRemoveUsers`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/client/groups/client.go internal/provider/client/groups/client_test.go
git commit -m "feat(groups): add RemoveUsers client method"
```

---

## Task 5: Remove user_ids from Update method payload

**Files:**
- Modify: `internal/provider/client/groups/client.go:86-117`

**Step 1: Create temporary struct for update payload**

Replace the Update method implementation with:

```go
func (c *Client) Update(id string, group *Group) (*Group, error) {
	// Create update payload without user_ids (API ignores it)
	updatePayload := struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Permissions *GroupPermissions `json:"permissions,omitempty"`
	}{
		Name:        group.Name,
		Description: group.Description,
		Permissions: group.Permissions,
	}

	body, err := json.Marshal(updatePayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/update", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var updatedGroup Group
	if err := json.NewDecoder(resp.Body).Decode(&updatedGroup); err != nil {
		return nil, err
	}

	return &updatedGroup, nil
}
```

**Step 2: Verify tests still pass**

Run: `go test ./internal/provider/client/groups/... -v`
Expected: All tests PASS

**Step 3: Commit**

```bash
git add internal/provider/client/groups/client.go
git commit -m "fix(groups): remove user_ids from Update payload"
```

---

## Task 6: Update resource Create to use AddUsers

**Files:**
- Modify: `internal/provider/groups_resource.go:161-309`

**Step 1: Modify Create method to call AddUsers**

Update the Create method to orchestrate user addition:

```go
func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First, create the group with basic information
	createGroup := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	createdGroup, err := r.client.Create(createGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			fmt.Sprintf("Could not create group: %s", err),
		)
		return
	}

	// Add users if provided
	if !plan.UserIDs.IsNull() && len(plan.UserIDs.Elements()) > 0 {
		var userIDs []string
		diags = plan.UserIDs.ElementsAs(ctx, &userIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		err = r.client.AddUsers(createdGroup.ID, userIDs)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding users to group",
				fmt.Sprintf("Could not add users to group %s: %s", createdGroup.ID, err),
			)
			return
		}
	}

	// Now prepare the update with permissions
	updateGroup := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Handle permissions
	if !plan.Permissions.IsNull() {
		var permissions struct {
			Workspace struct {
				Models    bool `tfsdk:"models"`
				Knowledge bool `tfsdk:"knowledge"`
				Prompts   bool `tfsdk:"prompts"`
				Tools     bool `tfsdk:"tools"`
			} `tfsdk:"workspace"`
			Chat struct {
				FileUpload         bool `tfsdk:"file_upload"`
				Delete             bool `tfsdk:"delete"`
				Edit               bool `tfsdk:"edit"`
				Temporary          bool `tfsdk:"temporary"`
				Controls           bool `tfsdk:"controls"`
				Valves             bool `tfsdk:"valves"`
				SystemPrompt       bool `tfsdk:"system_prompt"`
				Params             bool `tfsdk:"params"`
				DeleteMessage      bool `tfsdk:"delete_message"`
				ContinueResponse   bool `tfsdk:"continue_response"`
				RegenerateResponse bool `tfsdk:"regenerate_response"`
				RateResponse       bool `tfsdk:"rate_response"`
				Share              bool `tfsdk:"share"`
				Export             bool `tfsdk:"export"`
				Stt                bool `tfsdk:"stt"`
				Tts                bool `tfsdk:"tts"`
				Call               bool `tfsdk:"call"`
				MultipleModels     bool `tfsdk:"multiple_models"`
				TemporaryEnforced  bool `tfsdk:"temporary_enforced"`
			} `tfsdk:"chat"`
			Sharing struct {
				PublicModels    bool `tfsdk:"public_models"`
				PublicKnowledge bool `tfsdk:"public_knowledge"`
				PublicPrompts   bool `tfsdk:"public_prompts"`
				PublicTools     bool `tfsdk:"public_tools"`
				PublicNotes     bool `tfsdk:"public_notes"`
			} `tfsdk:"sharing"`
			Features struct {
				DirectToolServers bool `tfsdk:"direct_tool_servers"`
				WebSearch         bool `tfsdk:"web_search"`
				ImageGeneration   bool `tfsdk:"image_generation"`
				CodeInterpreter   bool `tfsdk:"code_interpreter"`
				Notes             bool `tfsdk:"notes"`
			} `tfsdk:"features"`
		}
		diags = plan.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateGroup.Permissions = &groups.GroupPermissions{
			Workspace: groups.WorkspacePermissions{
				Models:    permissions.Workspace.Models,
				Knowledge: permissions.Workspace.Knowledge,
				Prompts:   permissions.Workspace.Prompts,
				Tools:     permissions.Workspace.Tools,
			},
			Chat: groups.ChatPermissions{
				FileUpload:         permissions.Chat.FileUpload,
				Delete:             permissions.Chat.Delete,
				Edit:               permissions.Chat.Edit,
				Temporary:          permissions.Chat.Temporary,
				Controls:           permissions.Chat.Controls,
				Valves:             permissions.Chat.Valves,
				SystemPrompt:       permissions.Chat.SystemPrompt,
				Params:             permissions.Chat.Params,
				DeleteMessage:      permissions.Chat.DeleteMessage,
				ContinueResponse:   permissions.Chat.ContinueResponse,
				RegenerateResponse: permissions.Chat.RegenerateResponse,
				RateResponse:       permissions.Chat.RateResponse,
				Share:              permissions.Chat.Share,
				Export:             permissions.Chat.Export,
				Stt:                permissions.Chat.Stt,
				Tts:                permissions.Chat.Tts,
				Call:               permissions.Chat.Call,
				MultipleModels:     permissions.Chat.MultipleModels,
				TemporaryEnforced:  permissions.Chat.TemporaryEnforced,
			},
			Sharing: groups.SharingPermissions{
				PublicModels:    permissions.Sharing.PublicModels,
				PublicKnowledge: permissions.Sharing.PublicKnowledge,
				PublicPrompts:   permissions.Sharing.PublicPrompts,
				PublicTools:     permissions.Sharing.PublicTools,
				PublicNotes:     permissions.Sharing.PublicNotes,
			},
			Features: groups.FeaturesPermissions{
				DirectToolServers: permissions.Features.DirectToolServers,
				WebSearch:         permissions.Features.WebSearch,
				ImageGeneration:   permissions.Features.ImageGeneration,
				CodeInterpreter:   permissions.Features.CodeInterpreter,
				Notes:             permissions.Features.Notes,
			},
		}
	}

	// Update the group with permissions
	updatedGroup, err := r.client.Update(createdGroup.ID, updateGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			fmt.Sprintf("Could not update group with ID %s: %s", createdGroup.ID, err),
		)
		return
	}

	plan.ID = types.StringValue(updatedGroup.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
```

**Step 2: Verify code compiles**

Run: `go build`
Expected: Successful compilation

**Step 3: Commit**

```bash
git add internal/provider/groups_resource.go
git commit -m "feat(groups): update Create to use AddUsers"
```

---

## Task 7: Update resource Read to use GetUsers

**Files:**
- Modify: `internal/provider/groups_resource.go:311-483`

**Step 1: Modify Read method to call GetUsers**

Update the Read method to fetch users separately:

```go
func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			fmt.Sprintf("Could not read group ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.Name = types.StringValue(group.Name)
	state.Description = types.StringValue(group.Description)

	// Get users separately
	users, err := r.client.GetUsers(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group users",
			fmt.Sprintf("Could not read users for group ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Extract user IDs
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	sort.Strings(userIDs)
	userIDsList, diags := types.ListValueFrom(ctx, types.StringType, userIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UserIDs = userIDsList

	if group.Permissions != nil {
		workspaceAttrs := map[string]attr.Value{
			"models":    types.BoolValue(group.Permissions.Workspace.Models),
			"knowledge": types.BoolValue(group.Permissions.Workspace.Knowledge),
			"prompts":   types.BoolValue(group.Permissions.Workspace.Prompts),
			"tools":     types.BoolValue(group.Permissions.Workspace.Tools),
		}

		chatAttrs := map[string]attr.Value{
			"file_upload":         types.BoolValue(group.Permissions.Chat.FileUpload),
			"delete":              types.BoolValue(group.Permissions.Chat.Delete),
			"edit":                types.BoolValue(group.Permissions.Chat.Edit),
			"temporary":           types.BoolValue(group.Permissions.Chat.Temporary),
			"controls":            types.BoolValue(group.Permissions.Chat.Controls),
			"valves":              types.BoolValue(group.Permissions.Chat.Valves),
			"system_prompt":       types.BoolValue(group.Permissions.Chat.SystemPrompt),
			"params":              types.BoolValue(group.Permissions.Chat.Params),
			"delete_message":      types.BoolValue(group.Permissions.Chat.DeleteMessage),
			"continue_response":   types.BoolValue(group.Permissions.Chat.ContinueResponse),
			"regenerate_response": types.BoolValue(group.Permissions.Chat.RegenerateResponse),
			"rate_response":       types.BoolValue(group.Permissions.Chat.RateResponse),
			"share":               types.BoolValue(group.Permissions.Chat.Share),
			"export":              types.BoolValue(group.Permissions.Chat.Export),
			"stt":                 types.BoolValue(group.Permissions.Chat.Stt),
			"tts":                 types.BoolValue(group.Permissions.Chat.Tts),
			"call":                types.BoolValue(group.Permissions.Chat.Call),
			"multiple_models":     types.BoolValue(group.Permissions.Chat.MultipleModels),
			"temporary_enforced":  types.BoolValue(group.Permissions.Chat.TemporaryEnforced),
		}

		sharingAttrs := map[string]attr.Value{
			"public_models":    types.BoolValue(group.Permissions.Sharing.PublicModels),
			"public_knowledge": types.BoolValue(group.Permissions.Sharing.PublicKnowledge),
			"public_prompts":   types.BoolValue(group.Permissions.Sharing.PublicPrompts),
			"public_tools":     types.BoolValue(group.Permissions.Sharing.PublicTools),
			"public_notes":     types.BoolValue(group.Permissions.Sharing.PublicNotes),
		}

		featuresAttrs := map[string]attr.Value{
			"direct_tool_servers": types.BoolValue(group.Permissions.Features.DirectToolServers),
			"web_search":          types.BoolValue(group.Permissions.Features.WebSearch),
			"image_generation":    types.BoolValue(group.Permissions.Features.ImageGeneration),
			"code_interpreter":    types.BoolValue(group.Permissions.Features.CodeInterpreter),
			"notes":               types.BoolValue(group.Permissions.Features.Notes),
		}

		workspaceObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"models":    types.BoolType,
				"knowledge": types.BoolType,
				"prompts":   types.BoolType,
				"tools":     types.BoolType,
			},
			workspaceAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		chatObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"file_upload":         types.BoolType,
				"delete":              types.BoolType,
				"edit":                types.BoolType,
				"temporary":           types.BoolType,
				"controls":            types.BoolType,
				"valves":              types.BoolType,
				"system_prompt":       types.BoolType,
				"params":              types.BoolType,
				"delete_message":      types.BoolType,
				"continue_response":   types.BoolType,
				"regenerate_response": types.BoolType,
				"rate_response":       types.BoolType,
				"share":               types.BoolType,
				"export":              types.BoolType,
				"stt":                 types.BoolType,
				"tts":                 types.BoolType,
				"call":                types.BoolType,
				"multiple_models":     types.BoolType,
				"temporary_enforced":  types.BoolType,
			},
			chatAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sharingObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"public_models":    types.BoolType,
				"public_knowledge": types.BoolType,
				"public_prompts":   types.BoolType,
				"public_tools":     types.BoolType,
				"public_notes":     types.BoolType,
			},
			sharingAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		featuresObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"direct_tool_servers": types.BoolType,
				"web_search":          types.BoolType,
				"image_generation":    types.BoolType,
				"code_interpreter":    types.BoolType,
				"notes":               types.BoolType,
			},
			featuresAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		permissionsObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"workspace": types.ObjectType{AttrTypes: workspaceObj.AttributeTypes(ctx)},
				"chat":      types.ObjectType{AttrTypes: chatObj.AttributeTypes(ctx)},
				"sharing":   types.ObjectType{AttrTypes: sharingObj.AttributeTypes(ctx)},
				"features":  types.ObjectType{AttrTypes: featuresObj.AttributeTypes(ctx)},
			},
			map[string]attr.Value{
				"workspace": workspaceObj,
				"chat":      chatObj,
				"sharing":   sharingObj,
				"features":  featuresObj,
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Permissions = permissionsObj
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
```

**Step 2: Verify code compiles**

Run: `go build`
Expected: Successful compilation

**Step 3: Commit**

```bash
git add internal/provider/groups_resource.go
git commit -m "feat(groups): update Read to use GetUsers"
```

---

## Task 8: Update resource Update with surgical diff

**Files:**
- Modify: `internal/provider/groups_resource.go:485-614`

**Step 1: Implement Update method with user diffing**

Replace the Update method:

```go
func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current users to calculate diff
	currentUsers, err := r.client.GetUsers(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading current group users",
			fmt.Sprintf("Could not read current users for group ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Extract current user IDs
	currentUserIDs := make(map[string]bool)
	for _, user := range currentUsers {
		currentUserIDs[user.ID] = true
	}

	// Get planned user IDs
	var plannedUserIDs []string
	if !plan.UserIDs.IsNull() {
		diags = plan.UserIDs.ElementsAs(ctx, &plannedUserIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plannedUserIDsMap := make(map[string]bool)
	for _, id := range plannedUserIDs {
		plannedUserIDsMap[id] = true
	}

	// Calculate diff
	var toRemove []string
	var toAdd []string

	for id := range currentUserIDs {
		if !plannedUserIDsMap[id] {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range plannedUserIDs {
		if !currentUserIDs[id] {
			toAdd = append(toAdd, id)
		}
	}

	// Remove users first (safer to remove before adding)
	if len(toRemove) > 0 {
		err = r.client.RemoveUsers(plan.ID.ValueString(), toRemove)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing users from group",
				fmt.Sprintf("Could not remove users from group ID %s: %s", plan.ID.ValueString(), err),
			)
			return
		}
	}

	// Add new users
	if len(toAdd) > 0 {
		err = r.client.AddUsers(plan.ID.ValueString(), toAdd)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding users to group",
				fmt.Sprintf("Could not add users to group ID %s: %s", plan.ID.ValueString(), err),
			)
			return
		}
	}

	// Update group metadata (name, description, permissions)
	group := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Permissions.IsNull() {
		var permissions struct {
			Workspace struct {
				Models    bool `tfsdk:"models"`
				Knowledge bool `tfsdk:"knowledge"`
				Prompts   bool `tfsdk:"prompts"`
				Tools     bool `tfsdk:"tools"`
			} `tfsdk:"workspace"`
			Chat struct {
				FileUpload         bool `tfsdk:"file_upload"`
				Delete             bool `tfsdk:"delete"`
				Edit               bool `tfsdk:"edit"`
				Temporary          bool `tfsdk:"temporary"`
				Controls           bool `tfsdk:"controls"`
				Valves             bool `tfsdk:"valves"`
				SystemPrompt       bool `tfsdk:"system_prompt"`
				Params             bool `tfsdk:"params"`
				DeleteMessage      bool `tfsdk:"delete_message"`
				ContinueResponse   bool `tfsdk:"continue_response"`
				RegenerateResponse bool `tfsdk:"regenerate_response"`
				RateResponse       bool `tfsdk:"rate_response"`
				Share              bool `tfsdk:"share"`
				Export             bool `tfsdk:"export"`
				Stt                bool `tfsdk:"stt"`
				Tts                bool `tfsdk:"tts"`
				Call               bool `tfsdk:"call"`
				MultipleModels     bool `tfsdk:"multiple_models"`
				TemporaryEnforced  bool `tfsdk:"temporary_enforced"`
			} `tfsdk:"chat"`
			Sharing struct {
				PublicModels    bool `tfsdk:"public_models"`
				PublicKnowledge bool `tfsdk:"public_knowledge"`
				PublicPrompts   bool `tfsdk:"public_prompts"`
				PublicTools     bool `tfsdk:"public_tools"`
				PublicNotes     bool `tfsdk:"public_notes"`
			} `tfsdk:"sharing"`
			Features struct {
				DirectToolServers bool `tfsdk:"direct_tool_servers"`
				WebSearch         bool `tfsdk:"web_search"`
				ImageGeneration   bool `tfsdk:"image_generation"`
				CodeInterpreter   bool `tfsdk:"code_interpreter"`
				Notes             bool `tfsdk:"notes"`
			} `tfsdk:"features"`
		}
		diags = plan.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		group.Permissions = &groups.GroupPermissions{
			Workspace: groups.WorkspacePermissions{
				Models:    permissions.Workspace.Models,
				Knowledge: permissions.Workspace.Knowledge,
				Prompts:   permissions.Workspace.Prompts,
				Tools:     permissions.Workspace.Tools,
			},
			Chat: groups.ChatPermissions{
				FileUpload:         permissions.Chat.FileUpload,
				Delete:             permissions.Chat.Delete,
				Edit:               permissions.Chat.Edit,
				Temporary:          permissions.Chat.Temporary,
				Controls:           permissions.Chat.Controls,
				Valves:             permissions.Chat.Valves,
				SystemPrompt:       permissions.Chat.SystemPrompt,
				Params:             permissions.Chat.Params,
				DeleteMessage:      permissions.Chat.DeleteMessage,
				ContinueResponse:   permissions.Chat.ContinueResponse,
				RegenerateResponse: permissions.Chat.RegenerateResponse,
				RateResponse:       permissions.Chat.RateResponse,
				Share:              permissions.Chat.Share,
				Export:             permissions.Chat.Export,
				Stt:                permissions.Chat.Stt,
				Tts:                permissions.Chat.Tts,
				Call:               permissions.Chat.Call,
				MultipleModels:     permissions.Chat.MultipleModels,
				TemporaryEnforced:  permissions.Chat.TemporaryEnforced,
			},
			Sharing: groups.SharingPermissions{
				PublicModels:    permissions.Sharing.PublicModels,
				PublicKnowledge: permissions.Sharing.PublicKnowledge,
				PublicPrompts:   permissions.Sharing.PublicPrompts,
				PublicTools:     permissions.Sharing.PublicTools,
				PublicNotes:     permissions.Sharing.PublicNotes,
			},
			Features: groups.FeaturesPermissions{
				DirectToolServers: permissions.Features.DirectToolServers,
				WebSearch:         permissions.Features.WebSearch,
				ImageGeneration:   permissions.Features.ImageGeneration,
				CodeInterpreter:   permissions.Features.CodeInterpreter,
				Notes:             permissions.Features.Notes,
			},
		}
	}

	updatedGroup, err := r.client.Update(plan.ID.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			fmt.Sprintf("Could not update group ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	plan.ID = types.StringValue(updatedGroup.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
```

**Step 2: Verify code compiles**

Run: `go build`
Expected: Successful compilation

**Step 3: Commit**

```bash
git add internal/provider/groups_resource.go
git commit -m "feat(groups): update Update with surgical user diff"
```

---

## Task 9: Update groups data source to use GetUsers

**Files:**
- Modify: `internal/provider/groups_data_source.go`

**Step 1: Read the current groups data source implementation**

Run: `cat internal/provider/groups_data_source.go | grep -A 20 "func (d \*GroupsDataSource) Read"`

**Step 2: Update Read method to use GetUsers for each group**

Find the Read method and update the user ID fetching logic to call GetUsers:

```go
// In the loop where groups are processed, replace user_ids handling with:
users, err := d.client.GetUsers(group.ID)
if err != nil {
	resp.Diagnostics.AddError(
		"Error reading group users",
		fmt.Sprintf("Could not read users for group ID %s: %s", group.ID, err),
	)
	return
}

userIDs := make([]string, len(users))
for i, user := range users {
	userIDs[i] = user.ID
}
sort.Strings(userIDs)
```

**Step 3: Verify code compiles**

Run: `go build`
Expected: Successful compilation

**Step 4: Commit**

```bash
git add internal/provider/groups_data_source.go
git commit -m "feat(groups): update data source to use GetUsers"
```

---

## Task 10: Run all tests and verify

**Files:**
- N/A (testing phase)

**Step 1: Run client tests**

Run: `go test ./internal/provider/client/groups/... -v`
Expected: All tests PASS

**Step 2: Build provider**

Run: `go build`
Expected: Successful build with no errors

**Step 3: Run provider tests (if available)**

Run: `go test ./internal/provider/... -v`
Expected: All tests PASS

**Step 4: Final commit**

```bash
git add -A
git commit -m "chore: groups API migration complete"
```

---

## Testing Against Live API (Manual)

After implementation, test against a live Open WebUI instance:

1. Set environment variables:
   ```bash
   export OPENWEBUI_ENDPOINT="https://your-instance.com"
   export OPENWEBUI_TOKEN="your-token"
   ```

2. Create test Terraform config:
   ```hcl
   terraform {
     required_providers {
       openwebui = {
         source = "registry.terraform.io/insight2profit/openwebui"
       }
     }
   }

   resource "openwebui_group" "test" {
     name        = "Test Group"
     description = "API migration test"
     user_ids    = ["user-id-1", "user-id-2"]

     permissions {
       workspace {
         models    = false
         knowledge = false
         prompts   = false
         tools     = false
       }
       # ... rest of permissions
     }
   }
   ```

3. Test operations:
   - `terraform init`
   - `terraform plan` - should show creation
   - `terraform apply` - creates group with users
   - Modify `user_ids` in config
   - `terraform plan` - should show only user changes
   - `terraform apply` - should only modify users (surgical diff)
   - `terraform destroy` - cleans up

4. Verify in Open WebUI UI:
   - Group exists with correct users
   - Member count matches
   - Permissions are correct
