// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package tools

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Tool represents the Terraform schema model
type Tool struct {
	ID            types.String         `tfsdk:"id"`
	UserID        types.String         `tfsdk:"user_id"`
	Name          types.String         `tfsdk:"name"`
	Content       types.String         `tfsdk:"content"`
	Specs         jsontypes.Normalized `tfsdk:"specs"`
	Meta          *ToolMeta            `tfsdk:"meta"`
	AccessControl *AccessControl       `tfsdk:"access_control"`
	UpdatedAt     types.Int64          `tfsdk:"updated_at"`
	CreatedAt     types.Int64          `tfsdk:"created_at"`
}

// APITool represents the API response/request model
type APITool struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id,omitempty"`
	Name          string                 `json:"name"`
	Content       string                 `json:"content"`
	Specs         []map[string]interface{} `json:"specs,omitempty"`
	Meta          *APIToolMeta           `json:"meta"`
	AccessControl *APIAccessControl      `json:"access_control,omitempty"`
	UpdatedAt     int64                  `json:"updated_at,omitempty"`
	CreatedAt     int64                  `json:"created_at,omitempty"`
}

// ToolMeta holds tool metadata
type ToolMeta struct {
	Description types.String `tfsdk:"description"`
	Manifest    types.Map    `tfsdk:"manifest"`
}

// APIToolMeta represents the API tool metadata
type APIToolMeta struct {
	Description string                 `json:"description,omitempty"`
	Manifest    map[string]interface{} `json:"manifest,omitempty"`
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

// Helper function to convert API tool to Terraform tool
func APIToTool(apiTool *APITool) *Tool {
	tool := &Tool{
		ID:        types.StringValue(apiTool.ID),
		UserID:    types.StringValue(apiTool.UserID),
		Name:      types.StringValue(apiTool.Name),
		Content:   types.StringValue(apiTool.Content),
		UpdatedAt: types.Int64Value(apiTool.UpdatedAt),
		CreatedAt: types.Int64Value(apiTool.CreatedAt),
	}

	// Handle Specs - marshal to JSON string for jsontypes.Normalized
	if len(apiTool.Specs) > 0 {
		// jsontypes.Normalized will handle the JSON marshaling internally
		// We just need to store the raw JSON string
		specsJSON := ""
		if specsBytes, err := json.Marshal(apiTool.Specs); err == nil {
			specsJSON = string(specsBytes)
		}
		tool.Specs = jsontypes.NewNormalizedValue(specsJSON)
	}

	// Handle Meta
	if apiTool.Meta != nil {
		tool.Meta = &ToolMeta{}
		if apiTool.Meta.Description != "" {
			tool.Meta.Description = types.StringValue(apiTool.Meta.Description)
		}
		if len(apiTool.Meta.Manifest) > 0 {
			// Convert manifest map to types.Map
			manifestMap := make(map[string]attr.Value)
			for k, v := range apiTool.Meta.Manifest {
				// Convert interface{} to string representation
				manifestMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
			if manifestTypes, diags := types.MapValue(types.StringType, manifestMap); diags == nil {
				tool.Meta.Manifest = manifestTypes
			}
		}
	}

	// Handle AccessControl
	if apiTool.AccessControl != nil {
		tool.AccessControl = &AccessControl{}
		if apiTool.AccessControl.Read != nil {
			tool.AccessControl.Read = &AccessGroup{
				GroupIDs: make([]types.String, len(apiTool.AccessControl.Read.GroupIDs)),
				UserIDs:  make([]types.String, len(apiTool.AccessControl.Read.UserIDs)),
			}
			for i, id := range apiTool.AccessControl.Read.GroupIDs {
				tool.AccessControl.Read.GroupIDs[i] = types.StringValue(id)
			}
			for i, id := range apiTool.AccessControl.Read.UserIDs {
				tool.AccessControl.Read.UserIDs[i] = types.StringValue(id)
			}
		}
		if apiTool.AccessControl.Write != nil {
			tool.AccessControl.Write = &AccessGroup{
				GroupIDs: make([]types.String, len(apiTool.AccessControl.Write.GroupIDs)),
				UserIDs:  make([]types.String, len(apiTool.AccessControl.Write.UserIDs)),
			}
			for i, id := range apiTool.AccessControl.Write.GroupIDs {
				tool.AccessControl.Write.GroupIDs[i] = types.StringValue(id)
			}
			for i, id := range apiTool.AccessControl.Write.UserIDs {
				tool.AccessControl.Write.UserIDs[i] = types.StringValue(id)
			}
		}
	}

	return tool
}
