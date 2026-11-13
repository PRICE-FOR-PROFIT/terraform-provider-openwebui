// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/functions"
)

var (
	_ resource.Resource                = &FunctionResource{}
	_ resource.ResourceWithImportState = &FunctionResource{}
)

func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

type FunctionResource struct {
	client *functions.Client
}

func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *FunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	client, ok := clients["functions"].(*functions.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *functions.Client, got: %T. Please report this issue to the provider developers.", clients["functions"]),
		)
		return
	}

	r.client = client
}

func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a function in OpenWebUI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the function.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who created the function.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the function.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the function.",
				Optional:    true,
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/code of the function.",
				Required:    true,
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Function metadata.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "Description of the function.",
						Optional:    true,
					},
					"manifest": schema.MapAttribute{
						Description: "Function manifest metadata.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the function is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_global": schema.BoolAttribute{
				Description: "Whether the function is global.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.Int64Attribute{
				Description: "Timestamp when the function was created.",
				Computed:    true,
			},
			"updated_at": schema.Int64Attribute{
				Description: "Timestamp when the function was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functions.Function
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiFunction := &functions.APIFunction{
		ID:       plan.ID.ValueString(),
		Name:     plan.Name.ValueString(),
		Content:  plan.Content.ValueString(),
		IsActive: plan.IsActive.ValueBool(),
		IsGlobal: plan.IsGlobal.ValueBool(),
	}

	if !plan.Type.IsNull() {
		apiFunction.Type = plan.Type.ValueString()
	}

	// Handle Meta
	if plan.Meta != nil {
		apiFunction.Meta = &functions.APIFunctionMeta{}
		if !plan.Meta.Description.IsNull() {
			apiFunction.Meta.Description = plan.Meta.Description.ValueString()
		}
		if !plan.Meta.Manifest.IsNull() {
			manifestMap := make(map[string]interface{})
			diags := plan.Meta.Manifest.ElementsAs(ctx, &manifestMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiFunction.Meta.Manifest = manifestMap
		}
	}

	function, err := r.client.Create(apiFunction)
	if err != nil {
		resp.Diagnostics.AddError("Error creating function", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := functions.APIToFunction(function)

	// Ensure the ID is set in the state
	if state.ID.IsNull() {
		resp.Diagnostics.AddError("Error creating function", "Function ID is null after creation")
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functions.Function
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	function, err := r.client.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading function", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := functions.APIToFunction(function)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan functions.Function
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state functions.Function
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiFunction := &functions.APIFunction{
		ID:       plan.ID.ValueString(),
		Name:     plan.Name.ValueString(),
		Content:  plan.Content.ValueString(),
		IsActive: plan.IsActive.ValueBool(),
		IsGlobal: plan.IsGlobal.ValueBool(),
	}

	if !plan.Type.IsNull() {
		apiFunction.Type = plan.Type.ValueString()
	}

	// Handle Meta
	if plan.Meta != nil {
		apiFunction.Meta = &functions.APIFunctionMeta{}
		if !plan.Meta.Description.IsNull() {
			apiFunction.Meta.Description = plan.Meta.Description.ValueString()
		}
		if !plan.Meta.Manifest.IsNull() {
			manifestMap := make(map[string]interface{})
			diags := plan.Meta.Manifest.ElementsAs(ctx, &manifestMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiFunction.Meta.Manifest = manifestMap
		}
	}

	function, err := r.client.Update(state.ID.ValueString(), apiFunction)
	if err != nil {
		resp.Diagnostics.AddError("Error updating function", err.Error())
		return
	}

	// Convert API response back to Terraform model
	newState := functions.APIToFunction(function)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functions.Function
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting function", err.Error())
		return
	}
}

func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
