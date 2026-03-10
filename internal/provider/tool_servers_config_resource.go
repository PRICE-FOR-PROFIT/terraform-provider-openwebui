// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/configs"
)

var (
	_ resource.Resource                = &ToolServersConfigResource{}
	_ resource.ResourceWithImportState = &ToolServersConfigResource{}
)

func NewToolServersConfigResource() resource.Resource {
	return &ToolServersConfigResource{}
}

type ToolServersConfigResource struct {
	client *configs.Client
}

func (r *ToolServersConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool_servers_config"
}

func (r *ToolServersConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected map[string]interface{}, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	client, ok := clients["configs"].(*configs.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *configs.Client, got: %T. Please report this issue to the provider developers.", clients["configs"]),
		)
		return
	}

	r.client = client
}

func (r *ToolServersConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWebUI tool servers configuration. This is a singleton resource with a fixed ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Fixed identifier for the tool servers config (always 'tool_servers').",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tool_server_connections": schema.ListNestedAttribute{
				Description: "List of tool server connections.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Description: "The URL of the tool server.",
							Required:    true,
						},
						"path": schema.StringAttribute{
							Description: "The path on the tool server.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the tool server.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("openapi"),
						},
						"auth_type": schema.StringAttribute{
							Description: "The authentication type.",
							Optional:    true,
						},
						"key": schema.StringAttribute{
							Description: "The authentication key.",
							Optional:    true,
							Sensitive:   true,
						},
						"config": schema.MapAttribute{
							Description: "Additional configuration for the tool server.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *ToolServersConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configs.ToolServersConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIToolServersConfig{
		ToolServerConnections: make([]configs.APIToolServerConnection, len(plan.ToolServerConnections)),
	}

	for i, conn := range plan.ToolServerConnections {
		apiConn := configs.APIToolServerConnection{
			URL:  conn.URL.ValueString(),
			Path: conn.Path.ValueString(),
		}
		if !conn.Type.IsNull() {
			apiConn.Type = conn.Type.ValueString()
		}
		if !conn.AuthType.IsNull() {
			apiConn.AuthType = conn.AuthType.ValueString()
		}
		if !conn.Key.IsNull() {
			apiConn.Key = conn.Key.ValueString()
		}
		if !conn.Config.IsNull() {
			configMap := make(map[string]interface{})
			diags := conn.Config.ElementsAs(ctx, &configMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiConn.Config = configMap
		}
		apiConfig.ToolServerConnections[i] = apiConn
	}

	config, err := r.client.UpdateToolServers(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error creating tool servers config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToToolServersConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolServersConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configs.ToolServersConfig
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetToolServers()
	if err != nil {
		resp.Diagnostics.AddError("Error reading tool servers config", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := configs.APIToToolServersConfig(config)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolServersConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configs.ToolServersConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIToolServersConfig{
		ToolServerConnections: make([]configs.APIToolServerConnection, len(plan.ToolServerConnections)),
	}

	for i, conn := range plan.ToolServerConnections {
		apiConn := configs.APIToolServerConnection{
			URL:  conn.URL.ValueString(),
			Path: conn.Path.ValueString(),
		}
		if !conn.Type.IsNull() {
			apiConn.Type = conn.Type.ValueString()
		}
		if !conn.AuthType.IsNull() {
			apiConn.AuthType = conn.AuthType.ValueString()
		}
		if !conn.Key.IsNull() {
			apiConn.Key = conn.Key.ValueString()
		}
		if !conn.Config.IsNull() {
			configMap := make(map[string]interface{})
			diags := conn.Config.ElementsAs(ctx, &configMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiConn.Config = configMap
		}
		apiConfig.ToolServerConnections[i] = apiConn
	}

	config, err := r.client.UpdateToolServers(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error updating tool servers config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToToolServersConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolServersConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Reset to empty array
	apiConfig := &configs.APIToolServersConfig{
		ToolServerConnections: []configs.APIToolServerConnection{},
	}

	_, err := r.client.UpdateToolServers(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting tool servers config", err.Error())
		return
	}
}

func (r *ToolServersConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID must be "tool_servers"
	if req.ID != "tool_servers" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be 'tool_servers', got: %s", req.ID),
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
