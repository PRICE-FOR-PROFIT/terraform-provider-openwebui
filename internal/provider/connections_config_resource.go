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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"terraform-provider-openwebui/internal/provider/client/configs"
)

var (
	_ resource.Resource                = &ConnectionsConfigResource{}
	_ resource.ResourceWithImportState = &ConnectionsConfigResource{}
)

func NewConnectionsConfigResource() resource.Resource {
	return &ConnectionsConfigResource{}
}

type ConnectionsConfigResource struct {
	client *configs.Client
}

func (r *ConnectionsConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connections_config"
}

func (r *ConnectionsConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConnectionsConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWebUI connections configuration. This is a singleton resource with a fixed ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Fixed identifier for the connections config (always 'connections').",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enable_direct_connections": schema.BoolAttribute{
				Description: "Whether to enable direct connections.",
				Required:    true,
			},
			"enable_base_models_cache": schema.BoolAttribute{
				Description: "Whether to enable base models cache.",
				Required:    true,
			},
		},
	}
}

func (r *ConnectionsConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configs.ConnectionsConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIConnectionsConfig{
		EnableDirectConnections: plan.EnableDirectConnections.ValueBool(),
		EnableBaseModelsCache:   plan.EnableBaseModelsCache.ValueBool(),
	}

	config, err := r.client.UpdateConnections(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connections config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToConnectionsConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ConnectionsConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configs.ConnectionsConfig
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetConnections()
	if err != nil {
		resp.Diagnostics.AddError("Error reading connections config", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := configs.APIToConnectionsConfig(config)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *ConnectionsConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configs.ConnectionsConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIConnectionsConfig{
		EnableDirectConnections: plan.EnableDirectConnections.ValueBool(),
		EnableBaseModelsCache:   plan.EnableBaseModelsCache.ValueBool(),
	}

	config, err := r.client.UpdateConnections(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error updating connections config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToConnectionsConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ConnectionsConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Reset to default values
	apiConfig := &configs.APIConnectionsConfig{
		EnableDirectConnections: false,
		EnableBaseModelsCache:   false,
	}

	_, err := r.client.UpdateConnections(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting connections config", err.Error())
		return
	}
}

func (r *ConnectionsConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID must be "connections"
	if req.ID != "connections" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be 'connections', got: %s", req.ID),
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
