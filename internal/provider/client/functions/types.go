// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package functions

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Function represents the Terraform schema model
type Function struct {
	ID        types.String   `tfsdk:"id"`
	UserID    types.String   `tfsdk:"user_id"`
	Name      types.String   `tfsdk:"name"`
	Type      types.String   `tfsdk:"type"`
	Content   types.String   `tfsdk:"content"`
	Meta      *FunctionMeta  `tfsdk:"meta"`
	IsActive  types.Bool     `tfsdk:"is_active"`
	IsGlobal  types.Bool     `tfsdk:"is_global"`
	UpdatedAt types.Int64    `tfsdk:"updated_at"`
	CreatedAt types.Int64    `tfsdk:"created_at"`
}

// APIFunction represents the API response/request model
type APIFunction struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id,omitempty"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type,omitempty"`
	Content   string                 `json:"content"`
	Meta      *APIFunctionMeta       `json:"meta"`
	IsActive  bool                   `json:"is_active,omitempty"`
	IsGlobal  bool                   `json:"is_global,omitempty"`
	UpdatedAt int64                  `json:"updated_at,omitempty"`
	CreatedAt int64                  `json:"created_at,omitempty"`
}

// FunctionMeta holds function metadata
type FunctionMeta struct {
	Description types.String `tfsdk:"description"`
	Manifest    types.Map    `tfsdk:"manifest"`
}

// APIFunctionMeta represents the API function metadata
type APIFunctionMeta struct {
	Description string                 `json:"description,omitempty"`
	Manifest    map[string]interface{} `json:"manifest,omitempty"`
}

// Helper function to convert API function to Terraform function
func APIToFunction(apiFunction *APIFunction) *Function {
	function := &Function{
		ID:        types.StringValue(apiFunction.ID),
		UserID:    types.StringValue(apiFunction.UserID),
		Name:      types.StringValue(apiFunction.Name),
		Type:      types.StringValue(apiFunction.Type),
		Content:   types.StringValue(apiFunction.Content),
		IsActive:  types.BoolValue(apiFunction.IsActive),
		IsGlobal:  types.BoolValue(apiFunction.IsGlobal),
		UpdatedAt: types.Int64Value(apiFunction.UpdatedAt),
		CreatedAt: types.Int64Value(apiFunction.CreatedAt),
	}

	// Handle Meta
	if apiFunction.Meta != nil {
		function.Meta = &FunctionMeta{}
		if apiFunction.Meta.Description != "" {
			function.Meta.Description = types.StringValue(apiFunction.Meta.Description)
		}
		if len(apiFunction.Meta.Manifest) > 0 {
			// Convert manifest map to types.Map
			manifestMap := make(map[string]attr.Value)
			for k, v := range apiFunction.Meta.Manifest {
				// Convert interface{} to string representation
				manifestMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
			if manifestTypes, diags := types.MapValue(types.StringType, manifestMap); diags == nil {
				function.Meta.Manifest = manifestTypes
			}
		}
	}

	return function
}
