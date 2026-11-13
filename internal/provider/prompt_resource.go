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

	"terraform-provider-openwebui/internal/provider/client/prompts"
)

var (
	_ resource.Resource                = &PromptResource{}
	_ resource.ResourceWithImportState = &PromptResource{}
)

func NewPromptResource() resource.Resource {
	return &PromptResource{}
}

type PromptResource struct {
	client *prompts.Client
}

func (r *PromptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *PromptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	client, ok := clients["prompts"].(*prompts.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *prompts.Client, got: %T. Please report this issue to the provider developers.", clients["prompts"]),
		)
		return
	}

	r.client = client
}

func (r *PromptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a prompt in OpenWebUI. The prompt command serves as the unique identifier.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The command/ID of the prompt (e.g., '/summarize'). This is the primary identifier and cannot be changed after creation.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who created the prompt.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the prompt.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/template of the prompt.",
				Required:    true,
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
			"timestamp": schema.Int64Attribute{
				Description: "Timestamp when the prompt was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *PromptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan prompts.Prompt
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model (id -> command)
	apiPrompt := &prompts.APIPrompt{
		Command: plan.ID.ValueString(), // Map id to command
		Title:   plan.Title.ValueString(),
		Content: plan.Content.ValueString(),
	}

	// Handle AccessControl
	if plan.AccessControl != nil {
		apiPrompt.AccessControl = &prompts.APIAccessControl{}
		if plan.AccessControl.Read != nil {
			apiPrompt.AccessControl.Read = &prompts.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Read.GroupIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Read.GroupIDs = append(apiPrompt.AccessControl.Read.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Read.UserIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Read.UserIDs = append(apiPrompt.AccessControl.Read.UserIDs, id.ValueString())
				}
			}
		}
		if plan.AccessControl.Write != nil {
			apiPrompt.AccessControl.Write = &prompts.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Write.GroupIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Write.GroupIDs = append(apiPrompt.AccessControl.Write.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Write.UserIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Write.UserIDs = append(apiPrompt.AccessControl.Write.UserIDs, id.ValueString())
				}
			}
		}
	}

	prompt, err := r.client.Create(apiPrompt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating prompt", err.Error())
		return
	}

	// Convert API response back to Terraform model
	state := prompts.APIToPrompt(prompt)

	// Ensure the ID is set in the state
	if state.ID.IsNull() {
		resp.Diagnostics.AddError("Error creating prompt", "Prompt ID is null after creation")
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PromptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state prompts.Prompt
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use id as command
	prompt, err := r.client.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading prompt", err.Error())
		return
	}

	// Convert API response to Terraform model
	newState := prompts.APIToPrompt(prompt)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *PromptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan prompts.Prompt
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state prompts.Prompt
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model (id -> command)
	apiPrompt := &prompts.APIPrompt{
		Command: plan.ID.ValueString(), // Map id to command
		Title:   plan.Title.ValueString(),
		Content: plan.Content.ValueString(),
	}

	// Handle AccessControl
	if plan.AccessControl != nil {
		apiPrompt.AccessControl = &prompts.APIAccessControl{}
		if plan.AccessControl.Read != nil {
			apiPrompt.AccessControl.Read = &prompts.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Read.GroupIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Read.GroupIDs = append(apiPrompt.AccessControl.Read.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Read.UserIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Read.UserIDs = append(apiPrompt.AccessControl.Read.UserIDs, id.ValueString())
				}
			}
		}
		if plan.AccessControl.Write != nil {
			apiPrompt.AccessControl.Write = &prompts.APIAccessGroup{
				GroupIDs: make([]string, 0),
				UserIDs:  make([]string, 0),
			}
			for _, id := range plan.AccessControl.Write.GroupIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Write.GroupIDs = append(apiPrompt.AccessControl.Write.GroupIDs, id.ValueString())
				}
			}
			for _, id := range plan.AccessControl.Write.UserIDs {
				if !id.IsNull() {
					apiPrompt.AccessControl.Write.UserIDs = append(apiPrompt.AccessControl.Write.UserIDs, id.ValueString())
				}
			}
		}
	}

	// Use state ID (which is the command) for the update
	prompt, err := r.client.Update(state.ID.ValueString(), apiPrompt)
	if err != nil {
		resp.Diagnostics.AddError("Error updating prompt", err.Error())
		return
	}

	// Convert API response back to Terraform model
	newState := prompts.APIToPrompt(prompt)

	// Ensure the ID is preserved
	if newState.ID.IsNull() {
		newState.ID = state.ID
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *PromptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state prompts.Prompt
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use id as command
	err := r.client.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting prompt", err.Error())
		return
	}
}

func (r *PromptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import uses the command as the ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
