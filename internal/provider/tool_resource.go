// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/tools"
)

var (
	_ resource.Resource                = &ToolResource{}
	_ resource.ResourceWithImportState = &ToolResource{}
)

func NewToolResource() resource.Resource {
	return &ToolResource{}
}

type ToolResource struct {
	client *tools.Client
}

func (r *ToolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (r *ToolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	client, ok := clients["tools"].(*tools.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tools.Client, got: %T. Please report this issue to the provider developers.", clients["tools"]),
		)
		return
	}

	r.client = client
}

func (r *ToolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a tool in OpenWebUI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the tool.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who created the tool.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the tool.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/code of the tool.",
				Required:    true,
			},
			"specs": schema.StringAttribute{
				Description: "Tool specifications as JSON. Handles arbitrary JSON structure with semantic equality (ignores whitespace/ordering differences).",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Tool metadata.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "Description of the tool.",
						Optional:    true,
					},
					"manifest": schema.MapAttribute{
						Description: "Tool manifest metadata.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"access_control": schema.SingleNestedAttribute{
				Description: "Access control settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"read": schema.SingleNestedAttribute{
						Description: "Read access settings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"group_ids": schema.ListAttribute{
								Description: "List of group IDs with read access.",
								Optional:    true,
								ElementType: types.StringType,
							},
							"user_ids": schema.ListAttribute{
								Description: "List of user IDs with read access.",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
					"write": schema.SingleNestedAttribute{
						Description: "Write access settings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"group_ids": schema.ListAttribute{
								Description: "List of group IDs with write access.",
								Optional:    true,
								ElementType: types.StringType,
							},
							"user_ids": schema.ListAttribute{
								Description: "List of user IDs with write access.",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"created_at": schema.Int64Attribute{
				Description: "Timestamp when the tool was created.",
				Computed:    true,
			},
			"updated_at": schema.Int64Attribute{
				Description: "Timestamp when the tool was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ToolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tools.Tool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiTool := &tools.APITool{
		ID:      plan.ID.ValueString(),
		Name:    plan.Name.ValueString(),
		Content: plan.Content.ValueString(),
	}

	// Handle Specs
	if !plan.Specs.IsNull() && !plan.Specs.IsUnknown() {
		var specs []map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Specs.ValueString()), &specs); err != nil {
			resp.Diagnostics.AddError("Error parsing specs", err.Error())
			return
		}
		apiTool.Specs = specs
	}

	// Handle Meta
	if plan.Meta != nil {
		apiTool.Meta = &tools.APIToolMeta{}
		if !plan.Meta.Description.IsNull() {
			apiTool.Meta.Description = plan.Meta.Description.ValueString()
		}
		if !plan.Meta.Manifest.IsNull() {
			manifestMap := make(map[string]interface{})
			diags := plan.Meta.Manifest.ElementsAs(ctx, &manifestMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiTool.Meta.Manifest = manifestMap
		}
	}

	// Handle AccessControl
	if plan.AccessControl != nil {
		apiTool.AccessControl = &tools.APIAccessControl{}
		if plan.AccessControl.Read != nil {
			apiTool.AccessControl.Read = &tools.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Read.GroupIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Read.GroupIDs = append(apiTool.AccessControl.Read.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Read.UserIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Read.UserIDs = append(apiTool.AccessControl.Read.UserIDs, id.ValueString())
				}
			}
		}
		if plan.AccessControl.Write != nil {
			apiTool.AccessControl.Write = &tools.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Write.GroupIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Write.GroupIDs = append(apiTool.AccessControl.Write.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Write.UserIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Write.UserIDs = append(apiTool.AccessControl.Write.UserIDs, id.ValueString())
				}
			}
		}
	}

	tool, err := r.client.Create(apiTool)
	if err != nil {
		resp.Diagnostics.AddError("Error creating tool", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := tools.APIToTool(tool)

	// Ensure the ID is set in the state
	if state.ID.IsNull() {
		resp.Diagnostics.AddError("Error creating tool", "Tool ID is null after creation")
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tools.Tool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tool, err := r.client.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading tool", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := tools.APIToTool(tool)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tools.Tool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state tools.Tool
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiTool := &tools.APITool{
		ID:      plan.ID.ValueString(),
		Name:    plan.Name.ValueString(),
		Content: plan.Content.ValueString(),
	}

	// Handle Specs
	if !plan.Specs.IsNull() && !plan.Specs.IsUnknown() {
		var specs []map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Specs.ValueString()), &specs); err != nil {
			resp.Diagnostics.AddError("Error parsing specs", err.Error())
			return
		}
		apiTool.Specs = specs
	}

	// Handle Meta
	if plan.Meta != nil {
		apiTool.Meta = &tools.APIToolMeta{}
		if !plan.Meta.Description.IsNull() {
			apiTool.Meta.Description = plan.Meta.Description.ValueString()
		}
		if !plan.Meta.Manifest.IsNull() {
			manifestMap := make(map[string]interface{})
			diags := plan.Meta.Manifest.ElementsAs(ctx, &manifestMap, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			apiTool.Meta.Manifest = manifestMap
		}
	}

	// Handle AccessControl
	if plan.AccessControl != nil {
		apiTool.AccessControl = &tools.APIAccessControl{}
		if plan.AccessControl.Read != nil {
			apiTool.AccessControl.Read = &tools.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Read.GroupIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Read.GroupIDs = append(apiTool.AccessControl.Read.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Read.UserIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Read.UserIDs = append(apiTool.AccessControl.Read.UserIDs, id.ValueString())
				}
			}
		}
		if plan.AccessControl.Write != nil {
			apiTool.AccessControl.Write = &tools.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Write.GroupIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Write.GroupIDs = append(apiTool.AccessControl.Write.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Write.UserIDs {
				if !id.IsNull() {
					apiTool.AccessControl.Write.UserIDs = append(apiTool.AccessControl.Write.UserIDs, id.ValueString())
				}
			}
		}
	}

	tool, err := r.client.Update(state.ID.ValueString(), apiTool)
	if err != nil {
		resp.Diagnostics.AddError("Error updating tool", err.Error())
		return
	}

	// Convert API response back to Terraform model
	newState := tools.APIToTool(tool)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *ToolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tools.Tool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting tool", err.Error())
		return
	}
}

func (r *ToolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
