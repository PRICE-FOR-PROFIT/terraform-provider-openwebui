// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package prompts

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Prompt represents the Terraform schema model
// Note: The API uses "command" as the primary key, which we map to "id" in Terraform
type Prompt struct {
	ID            types.String   `tfsdk:"id"` // Maps to API's "command" field
	UserID        types.String   `tfsdk:"user_id"`
	Title         types.String   `tfsdk:"title"`
	Content       types.String   `tfsdk:"content"`
	AccessControl *AccessControl `tfsdk:"access_control"`
	Timestamp     types.Int64    `tfsdk:"timestamp"`
}

// APIPrompt represents the API response/request model
type APIPrompt struct {
	Command       string            `json:"command"` // Primary key in API
	UserID        string            `json:"user_id,omitempty"`
	Title         string            `json:"title"`
	Content       string            `json:"content"`
	AccessControl *APIAccessControl `json:"access_control,omitempty"`
	Timestamp     int64             `json:"timestamp,omitempty"`
}

// AccessControl represents access control settings
type AccessControl struct {
	Read  *AccessGroup `tfsdk:"read"`
	Write *AccessGroup `tfsdk:"write"`
}

// APIAccessControl represents the API access control
type APIAccessControl struct {
	Read  *APIAccessGroup `json:"read,omitempty"`
	Write *APIAccessGroup `json:"write,omitempty"`
}

// AccessGroup represents a group of users/groups with access
type AccessGroup struct {
	GroupIDs []types.String `tfsdk:"group_ids"`
	UserIDs  []types.String `tfsdk:"user_ids"`
}

// APIAccessGroup represents the API access group
type APIAccessGroup struct {
	GroupIDs []string `json:"group_ids,omitempty"`
	UserIDs  []string `json:"user_ids,omitempty"`
}

// Helper function to convert API prompt to Terraform prompt
func APIToPrompt(apiPrompt *APIPrompt) *Prompt {
	prompt := &Prompt{
		ID:        types.StringValue(apiPrompt.Command), // Map command to id
		UserID:    types.StringValue(apiPrompt.UserID),
		Title:     types.StringValue(apiPrompt.Title),
		Content:   types.StringValue(apiPrompt.Content),
		Timestamp: types.Int64Value(apiPrompt.Timestamp),
	}

	// Handle AccessControl
	if apiPrompt.AccessControl != nil {
		prompt.AccessControl = &AccessControl{}
		if apiPrompt.AccessControl.Read != nil {
			prompt.AccessControl.Read = &AccessGroup{
				GroupIDs: make([]types.String, len(apiPrompt.AccessControl.Read.GroupIDs)),
				UserIDs:  make([]types.String, len(apiPrompt.AccessControl.Read.UserIDs)),
			}
			for i, id := range apiPrompt.AccessControl.Read.GroupIDs {
				prompt.AccessControl.Read.GroupIDs[i] = types.StringValue(id)
			}
			for i, id := range apiPrompt.AccessControl.Read.UserIDs {
				prompt.AccessControl.Read.UserIDs[i] = types.StringValue(id)
			}
		}
		if apiPrompt.AccessControl.Write != nil {
			prompt.AccessControl.Write = &AccessGroup{
				GroupIDs: make([]types.String, len(apiPrompt.AccessControl.Write.GroupIDs)),
				UserIDs:  make([]types.String, len(apiPrompt.AccessControl.Write.UserIDs)),
			}
			for i, id := range apiPrompt.AccessControl.Write.GroupIDs {
				prompt.AccessControl.Write.GroupIDs[i] = types.StringValue(id)
			}
			for i, id := range apiPrompt.AccessControl.Write.UserIDs {
				prompt.AccessControl.Write.UserIDs[i] = types.StringValue(id)
			}
		}
	}

	return prompt
}
