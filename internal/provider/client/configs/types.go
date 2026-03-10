// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package configs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConnectionsConfig represents the Terraform schema model for connections config
type ConnectionsConfig struct {
	ID                       types.String `tfsdk:"id"`
	EnableDirectConnections  types.Bool   `tfsdk:"enable_direct_connections"`
	EnableBaseModelsCache    types.Bool   `tfsdk:"enable_base_models_cache"`
}

// APIConnectionsConfig represents the API response/request model
type APIConnectionsConfig struct {
	EnableDirectConnections bool `json:"ENABLE_DIRECT_CONNECTIONS"`
	EnableBaseModelsCache   bool `json:"ENABLE_BASE_MODELS_CACHE"`
}

// ToolServersConfig represents the Terraform schema model for tool servers config
type ToolServersConfig struct {
	ID                    types.String              `tfsdk:"id"`
	ToolServerConnections []ToolServerConnection    `tfsdk:"tool_server_connections"`
}

// APIToolServersConfig represents the API response/request model
type APIToolServersConfig struct {
	ToolServerConnections []APIToolServerConnection `json:"TOOL_SERVER_CONNECTIONS"`
}

// ToolServerConnection represents a single tool server connection
type ToolServerConnection struct {
	URL      types.String `tfsdk:"url"`
	Path     types.String `tfsdk:"path"`
	Type     types.String `tfsdk:"type"`
	AuthType types.String `tfsdk:"auth_type"`
	Key      types.String `tfsdk:"key"`
	Config   types.Map    `tfsdk:"config"`
}

// APIToolServerConnection represents the API tool server connection
type APIToolServerConnection struct {
	URL      string                 `json:"url"`
	Path     string                 `json:"path"`
	Type     string                 `json:"type,omitempty"`
	AuthType string                 `json:"auth_type,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ModelsConfig represents the Terraform schema model for models config
type ModelsConfig struct {
	ID             types.String   `tfsdk:"id"`
	DefaultModels  types.String   `tfsdk:"default_models"`
	ModelOrderList []types.String `tfsdk:"model_order_list"`
}

// APIModelsConfig represents the API response/request model
type APIModelsConfig struct {
	DefaultModels  string   `json:"DEFAULT_MODELS"`
	ModelOrderList []string `json:"MODEL_ORDER_LIST"`
}

// Helper function to convert API connections config to Terraform model
func APIToConnectionsConfig(apiConfig *APIConnectionsConfig) *ConnectionsConfig {
	return &ConnectionsConfig{
		ID:                      types.StringValue("connections"),
		EnableDirectConnections: types.BoolValue(apiConfig.EnableDirectConnections),
		EnableBaseModelsCache:   types.BoolValue(apiConfig.EnableBaseModelsCache),
	}
}

// Helper function to convert API tool servers config to Terraform model
func APIToToolServersConfig(apiConfig *APIToolServersConfig) *ToolServersConfig {
	config := &ToolServersConfig{
		ID:                    types.StringValue("tool_servers"),
		ToolServerConnections: make([]ToolServerConnection, len(apiConfig.ToolServerConnections)),
	}

	for i, conn := range apiConfig.ToolServerConnections {
		config.ToolServerConnections[i] = ToolServerConnection{
			URL:      types.StringValue(conn.URL),
			Path:     types.StringValue(conn.Path),
			Type:     types.StringValue(conn.Type),
			AuthType: types.StringValue(conn.AuthType),
			Key:      types.StringValue(conn.Key),
		}
		// Handle Config map conversion if needed
		if len(conn.Config) > 0 {
			// Convert to types.Map - simplified version, could be enhanced
			config.ToolServerConnections[i].Config = types.MapNull(types.StringType)
		}
	}

	return config
}

// Helper function to convert API models config to Terraform model
func APIToModelsConfig(apiConfig *APIModelsConfig) *ModelsConfig {
	config := &ModelsConfig{
		ID:             types.StringValue("models_config"),
		DefaultModels:  types.StringValue(apiConfig.DefaultModels),
		ModelOrderList: make([]types.String, len(apiConfig.ModelOrderList)),
	}

	for i, model := range apiConfig.ModelOrderList {
		config.ModelOrderList[i] = types.StringValue(model)
	}

	return config
}
