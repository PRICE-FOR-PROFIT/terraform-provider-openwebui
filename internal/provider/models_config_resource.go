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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/configs"
)

var (
	_ resource.Resource                = &ModelsConfigResource{}
	_ resource.ResourceWithImportState = &ModelsConfigResource{}
)

func NewModelsConfigResource() resource.Resource {
	return &ModelsConfigResource{}
}

type ModelsConfigResource struct {
	client *configs.Client
}

func (r *ModelsConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_models_config"
}

func (r *ModelsConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModelsConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OpenWebUI models configuration (global settings). This is a singleton resource with a fixed ID. Note: This manages global model settings, different from the openwebui_model resource which manages individual model instances.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Fixed identifier for the models config (always 'models_config').",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"default_models": schema.StringAttribute{
				Description: "Default models configuration.",
				Optional:    true,
			},
			"model_order_list": schema.ListAttribute{
				Description: "List of model IDs in preferred order.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ModelsConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configs.ModelsConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIModelsConfig{
		DefaultModels:  plan.DefaultModels.ValueString(),
		ModelOrderList: make([]string, 0),
	}

	for _, model := range plan.ModelOrderList {
		if !model.IsNull() {
			apiConfig.ModelOrderList = append(apiConfig.ModelOrderList, model.ValueString())
		}
	}

	config, err := r.client.UpdateModels(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error creating models config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToModelsConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelsConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configs.ModelsConfig
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetModels()
	if err != nil {
		resp.Diagnostics.AddError("Error reading models config", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := configs.APIToModelsConfig(config)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelsConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configs.ModelsConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := &configs.APIModelsConfig{
		DefaultModels:  plan.DefaultModels.ValueString(),
		ModelOrderList: make([]string, 0),
	}

	for _, model := range plan.ModelOrderList {
		if !model.IsNull() {
			apiConfig.ModelOrderList = append(apiConfig.ModelOrderList, model.ValueString())
		}
	}

	config, err := r.client.UpdateModels(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error updating models config", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := configs.APIToModelsConfig(config)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelsConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Reset to empty/null values
	apiConfig := &configs.APIModelsConfig{
		DefaultModels:  "",
		ModelOrderList: []string{},
	}

	_, err := r.client.UpdateModels(apiConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting models config", err.Error())
		return
	}
}

func (r *ModelsConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID must be "models_config"
	if req.ID != "models_config" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be 'models_config', got: %s", req.ID),
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
