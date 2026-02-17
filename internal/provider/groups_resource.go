// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-openwebui/internal/provider/client/groups"
)

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

type GroupResource struct {
	client *groups.Client
}

type GroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	UserIDs     types.List   `tfsdk:"user_ids"`
	Permissions types.Object `tfsdk:"permissions"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group in OpenWebUI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the group.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the group.",
				Optional:    true,
			},
			"user_ids": schema.ListAttribute{
				Description: "List of user IDs in the group.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"permissions": schema.SingleNestedAttribute{
				Description: "Permissions for the group.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"workspace": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"models":    schema.BoolAttribute{Required: true},
							"knowledge": schema.BoolAttribute{Required: true},
							"prompts":   schema.BoolAttribute{Required: true},
							"tools":     schema.BoolAttribute{Required: true},
						},
					},
					"chat": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"file_upload":         schema.BoolAttribute{Required: true},
							"delete":              schema.BoolAttribute{Required: true},
							"edit":                schema.BoolAttribute{Required: true},
							"temporary":           schema.BoolAttribute{Required: true},
							"controls":            schema.BoolAttribute{Required: true},
							"valves":              schema.BoolAttribute{Required: true},
							"system_prompt":       schema.BoolAttribute{Required: true},
							"params":              schema.BoolAttribute{Required: true},
							"delete_message":      schema.BoolAttribute{Required: true},
							"continue_response":   schema.BoolAttribute{Required: true},
							"regenerate_response": schema.BoolAttribute{Required: true},
							"rate_response":       schema.BoolAttribute{Required: true},
							"share":               schema.BoolAttribute{Required: true},
							"export":              schema.BoolAttribute{Required: true},
							"stt":                 schema.BoolAttribute{Required: true},
							"tts":                 schema.BoolAttribute{Required: true},
							"call":                schema.BoolAttribute{Required: true},
							"multiple_models":     schema.BoolAttribute{Required: true},
							"temporary_enforced":  schema.BoolAttribute{Required: true},
						},
					},
					"sharing": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"public_models":    schema.BoolAttribute{Required: true},
							"public_knowledge": schema.BoolAttribute{Required: true},
							"public_prompts":   schema.BoolAttribute{Required: true},
							"public_tools":     schema.BoolAttribute{Required: true},
							"public_notes":     schema.BoolAttribute{Required: true},
						},
					},
					"features": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"direct_tool_servers": schema.BoolAttribute{Required: true},
							"web_search":          schema.BoolAttribute{Required: true},
							"image_generation":    schema.BoolAttribute{Required: true},
							"code_interpreter":    schema.BoolAttribute{Required: true},
							"notes":               schema.BoolAttribute{Required: true},
						},
					},
				},
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	client, ok := clients["groups"].(*groups.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *groups.Client, got: %T. Please report this issue to the provider developers.", clients["groups"]),
		)
		return
	}

	r.client = client
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First, create the group with basic information
	createGroup := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	createdGroup, err := r.client.Create(createGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			fmt.Sprintf("Could not create group: %s", err),
		)
		return
	}

	// Now prepare the update with all the additional information
	updateGroup := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Handle user IDs
	var userIDs []string
	diags = plan.UserIDs.ElementsAs(ctx, &userIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateGroup.UserIDs = userIDs

	// Handle permissions
	if !plan.Permissions.IsNull() {
		var permissions struct {
			Workspace struct {
				Models    bool `tfsdk:"models"`
				Knowledge bool `tfsdk:"knowledge"`
				Prompts   bool `tfsdk:"prompts"`
				Tools     bool `tfsdk:"tools"`
			} `tfsdk:"workspace"`
			Chat struct {
				FileUpload         bool `tfsdk:"file_upload"`
				Delete             bool `tfsdk:"delete"`
				Edit               bool `tfsdk:"edit"`
				Temporary          bool `tfsdk:"temporary"`
				Controls           bool `tfsdk:"controls"`
				Valves             bool `tfsdk:"valves"`
				SystemPrompt       bool `tfsdk:"system_prompt"`
				Params             bool `tfsdk:"params"`
				DeleteMessage      bool `tfsdk:"delete_message"`
				ContinueResponse   bool `tfsdk:"continue_response"`
				RegenerateResponse bool `tfsdk:"regenerate_response"`
				RateResponse       bool `tfsdk:"rate_response"`
				Share              bool `tfsdk:"share"`
				Export             bool `tfsdk:"export"`
				Stt                bool `tfsdk:"stt"`
				Tts                bool `tfsdk:"tts"`
				Call               bool `tfsdk:"call"`
				MultipleModels     bool `tfsdk:"multiple_models"`
				TemporaryEnforced  bool `tfsdk:"temporary_enforced"`
			} `tfsdk:"chat"`
			Sharing struct {
				PublicModels    bool `tfsdk:"public_models"`
				PublicKnowledge bool `tfsdk:"public_knowledge"`
				PublicPrompts   bool `tfsdk:"public_prompts"`
				PublicTools     bool `tfsdk:"public_tools"`
				PublicNotes     bool `tfsdk:"public_notes"`
			} `tfsdk:"sharing"`
			Features struct {
				DirectToolServers bool `tfsdk:"direct_tool_servers"`
				WebSearch         bool `tfsdk:"web_search"`
				ImageGeneration   bool `tfsdk:"image_generation"`
				CodeInterpreter   bool `tfsdk:"code_interpreter"`
				Notes             bool `tfsdk:"notes"`
			} `tfsdk:"features"`
		}
		diags = plan.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateGroup.Permissions = &groups.GroupPermissions{
			Workspace: groups.WorkspacePermissions{
				Models:    permissions.Workspace.Models,
				Knowledge: permissions.Workspace.Knowledge,
				Prompts:   permissions.Workspace.Prompts,
				Tools:     permissions.Workspace.Tools,
			},
			Chat: groups.ChatPermissions{
				FileUpload:         permissions.Chat.FileUpload,
				Delete:             permissions.Chat.Delete,
				Edit:               permissions.Chat.Edit,
				Temporary:          permissions.Chat.Temporary,
				Controls:           permissions.Chat.Controls,
				Valves:             permissions.Chat.Valves,
				SystemPrompt:       permissions.Chat.SystemPrompt,
				Params:             permissions.Chat.Params,
				DeleteMessage:      permissions.Chat.DeleteMessage,
				ContinueResponse:   permissions.Chat.ContinueResponse,
				RegenerateResponse: permissions.Chat.RegenerateResponse,
				RateResponse:       permissions.Chat.RateResponse,
				Share:              permissions.Chat.Share,
				Export:             permissions.Chat.Export,
				Stt:                permissions.Chat.Stt,
				Tts:                permissions.Chat.Tts,
				Call:               permissions.Chat.Call,
				MultipleModels:     permissions.Chat.MultipleModels,
				TemporaryEnforced:  permissions.Chat.TemporaryEnforced,
			},
			Sharing: groups.SharingPermissions{
				PublicModels:    permissions.Sharing.PublicModels,
				PublicKnowledge: permissions.Sharing.PublicKnowledge,
				PublicPrompts:   permissions.Sharing.PublicPrompts,
				PublicTools:     permissions.Sharing.PublicTools,
				PublicNotes:     permissions.Sharing.PublicNotes,
			},
			Features: groups.FeaturesPermissions{
				DirectToolServers: permissions.Features.DirectToolServers,
				WebSearch:         permissions.Features.WebSearch,
				ImageGeneration:   permissions.Features.ImageGeneration,
				CodeInterpreter:   permissions.Features.CodeInterpreter,
				Notes:             permissions.Features.Notes,
			},
		}
	}

	// Update the group with all the information
	updatedGroup, err := r.client.Update(createdGroup.ID, updateGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			fmt.Sprintf("Could not update group with ID %s: %s", createdGroup.ID, err),
		)
		return
	}

	plan.ID = types.StringValue(updatedGroup.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			fmt.Sprintf("Could not read group ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.Name = types.StringValue(group.Name)
	state.Description = types.StringValue(group.Description)

	sort.Strings(group.UserIDs)
	userIDs, diags := types.ListValueFrom(ctx, types.StringType, group.UserIDs)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UserIDs = userIDs

	if group.Permissions != nil {
		workspaceAttrs := map[string]attr.Value{
			"models":    types.BoolValue(group.Permissions.Workspace.Models),
			"knowledge": types.BoolValue(group.Permissions.Workspace.Knowledge),
			"prompts":   types.BoolValue(group.Permissions.Workspace.Prompts),
			"tools":     types.BoolValue(group.Permissions.Workspace.Tools),
		}

		chatAttrs := map[string]attr.Value{
			"file_upload":         types.BoolValue(group.Permissions.Chat.FileUpload),
			"delete":              types.BoolValue(group.Permissions.Chat.Delete),
			"edit":                types.BoolValue(group.Permissions.Chat.Edit),
			"temporary":           types.BoolValue(group.Permissions.Chat.Temporary),
			"controls":            types.BoolValue(group.Permissions.Chat.Controls),
			"valves":              types.BoolValue(group.Permissions.Chat.Valves),
			"system_prompt":       types.BoolValue(group.Permissions.Chat.SystemPrompt),
			"params":              types.BoolValue(group.Permissions.Chat.Params),
			"delete_message":      types.BoolValue(group.Permissions.Chat.DeleteMessage),
			"continue_response":   types.BoolValue(group.Permissions.Chat.ContinueResponse),
			"regenerate_response": types.BoolValue(group.Permissions.Chat.RegenerateResponse),
			"rate_response":       types.BoolValue(group.Permissions.Chat.RateResponse),
			"share":               types.BoolValue(group.Permissions.Chat.Share),
			"export":              types.BoolValue(group.Permissions.Chat.Export),
			"stt":                 types.BoolValue(group.Permissions.Chat.Stt),
			"tts":                 types.BoolValue(group.Permissions.Chat.Tts),
			"call":                types.BoolValue(group.Permissions.Chat.Call),
			"multiple_models":     types.BoolValue(group.Permissions.Chat.MultipleModels),
			"temporary_enforced":  types.BoolValue(group.Permissions.Chat.TemporaryEnforced),
		}

		sharingAttrs := map[string]attr.Value{
			"public_models":    types.BoolValue(group.Permissions.Sharing.PublicModels),
			"public_knowledge": types.BoolValue(group.Permissions.Sharing.PublicKnowledge),
			"public_prompts":   types.BoolValue(group.Permissions.Sharing.PublicPrompts),
			"public_tools":     types.BoolValue(group.Permissions.Sharing.PublicTools),
			"public_notes":     types.BoolValue(group.Permissions.Sharing.PublicNotes),
		}

		featuresAttrs := map[string]attr.Value{
			"direct_tool_servers": types.BoolValue(group.Permissions.Features.DirectToolServers),
			"web_search":          types.BoolValue(group.Permissions.Features.WebSearch),
			"image_generation":    types.BoolValue(group.Permissions.Features.ImageGeneration),
			"code_interpreter":    types.BoolValue(group.Permissions.Features.CodeInterpreter),
			"notes":               types.BoolValue(group.Permissions.Features.Notes),
		}

		workspaceObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"models":    types.BoolType,
				"knowledge": types.BoolType,
				"prompts":   types.BoolType,
				"tools":     types.BoolType,
			},
			workspaceAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		chatObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"file_upload":         types.BoolType,
				"delete":              types.BoolType,
				"edit":                types.BoolType,
				"temporary":           types.BoolType,
				"controls":            types.BoolType,
				"valves":              types.BoolType,
				"system_prompt":       types.BoolType,
				"params":              types.BoolType,
				"delete_message":      types.BoolType,
				"continue_response":   types.BoolType,
				"regenerate_response": types.BoolType,
				"rate_response":       types.BoolType,
				"share":               types.BoolType,
				"export":              types.BoolType,
				"stt":                 types.BoolType,
				"tts":                 types.BoolType,
				"call":                types.BoolType,
				"multiple_models":     types.BoolType,
				"temporary_enforced":  types.BoolType,
			},
			chatAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sharingObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"public_models":    types.BoolType,
				"public_knowledge": types.BoolType,
				"public_prompts":   types.BoolType,
				"public_tools":     types.BoolType,
				"public_notes":     types.BoolType,
			},
			sharingAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		featuresObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"direct_tool_servers": types.BoolType,
				"web_search":          types.BoolType,
				"image_generation":    types.BoolType,
				"code_interpreter":    types.BoolType,
				"notes":               types.BoolType,
			},
			featuresAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		permissionsObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"workspace": types.ObjectType{AttrTypes: workspaceObj.AttributeTypes(ctx)},
				"chat":      types.ObjectType{AttrTypes: chatObj.AttributeTypes(ctx)},
				"sharing":   types.ObjectType{AttrTypes: sharingObj.AttributeTypes(ctx)},
				"features":  types.ObjectType{AttrTypes: featuresObj.AttributeTypes(ctx)},
			},
			map[string]attr.Value{
				"workspace": workspaceObj,
				"chat":      chatObj,
				"sharing":   sharingObj,
				"features":  featuresObj,
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Permissions = permissionsObj
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := &groups.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	var userIDs []string
	diags = plan.UserIDs.ElementsAs(ctx, &userIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	group.UserIDs = userIDs

	if !plan.Permissions.IsNull() {
		var permissions struct {
			Workspace struct {
				Models    bool `tfsdk:"models"`
				Knowledge bool `tfsdk:"knowledge"`
				Prompts   bool `tfsdk:"prompts"`
				Tools     bool `tfsdk:"tools"`
			} `tfsdk:"workspace"`
			Chat struct {
				FileUpload         bool `tfsdk:"file_upload"`
				Delete             bool `tfsdk:"delete"`
				Edit               bool `tfsdk:"edit"`
				Temporary          bool `tfsdk:"temporary"`
				Controls           bool `tfsdk:"controls"`
				Valves             bool `tfsdk:"valves"`
				SystemPrompt       bool `tfsdk:"system_prompt"`
				Params             bool `tfsdk:"params"`
				DeleteMessage      bool `tfsdk:"delete_message"`
				ContinueResponse   bool `tfsdk:"continue_response"`
				RegenerateResponse bool `tfsdk:"regenerate_response"`
				RateResponse       bool `tfsdk:"rate_response"`
				Share              bool `tfsdk:"share"`
				Export             bool `tfsdk:"export"`
				Stt                bool `tfsdk:"stt"`
				Tts                bool `tfsdk:"tts"`
				Call               bool `tfsdk:"call"`
				MultipleModels     bool `tfsdk:"multiple_models"`
				TemporaryEnforced  bool `tfsdk:"temporary_enforced"`
			} `tfsdk:"chat"`
			Sharing struct {
				PublicModels    bool `tfsdk:"public_models"`
				PublicKnowledge bool `tfsdk:"public_knowledge"`
				PublicPrompts   bool `tfsdk:"public_prompts"`
				PublicTools     bool `tfsdk:"public_tools"`
				PublicNotes     bool `tfsdk:"public_notes"`
			} `tfsdk:"sharing"`
			Features struct {
				DirectToolServers bool `tfsdk:"direct_tool_servers"`
				WebSearch         bool `tfsdk:"web_search"`
				ImageGeneration   bool `tfsdk:"image_generation"`
				CodeInterpreter   bool `tfsdk:"code_interpreter"`
				Notes             bool `tfsdk:"notes"`
			} `tfsdk:"features"`
		}
		diags = plan.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		group.Permissions = &groups.GroupPermissions{
			Workspace: groups.WorkspacePermissions{
				Models:    permissions.Workspace.Models,
				Knowledge: permissions.Workspace.Knowledge,
				Prompts:   permissions.Workspace.Prompts,
				Tools:     permissions.Workspace.Tools,
			},
			Chat: groups.ChatPermissions{
				FileUpload:         permissions.Chat.FileUpload,
				Delete:             permissions.Chat.Delete,
				Edit:               permissions.Chat.Edit,
				Temporary:          permissions.Chat.Temporary,
				Controls:           permissions.Chat.Controls,
				Valves:             permissions.Chat.Valves,
				SystemPrompt:       permissions.Chat.SystemPrompt,
				Params:             permissions.Chat.Params,
				DeleteMessage:      permissions.Chat.DeleteMessage,
				ContinueResponse:   permissions.Chat.ContinueResponse,
				RegenerateResponse: permissions.Chat.RegenerateResponse,
				RateResponse:       permissions.Chat.RateResponse,
				Share:              permissions.Chat.Share,
				Export:             permissions.Chat.Export,
				Stt:                permissions.Chat.Stt,
				Tts:                permissions.Chat.Tts,
				Call:               permissions.Chat.Call,
				MultipleModels:     permissions.Chat.MultipleModels,
				TemporaryEnforced:  permissions.Chat.TemporaryEnforced,
			},
			Sharing: groups.SharingPermissions{
				PublicModels:    permissions.Sharing.PublicModels,
				PublicKnowledge: permissions.Sharing.PublicKnowledge,
				PublicPrompts:   permissions.Sharing.PublicPrompts,
				PublicTools:     permissions.Sharing.PublicTools,
				PublicNotes:     permissions.Sharing.PublicNotes,
			},
			Features: groups.FeaturesPermissions{
				DirectToolServers: permissions.Features.DirectToolServers,
				WebSearch:         permissions.Features.WebSearch,
				ImageGeneration:   permissions.Features.ImageGeneration,
				CodeInterpreter:   permissions.Features.CodeInterpreter,
				Notes:             permissions.Features.Notes,
			},
		}
	}

	updatedGroup, err := r.client.Update(plan.ID.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			fmt.Sprintf("Could not update group ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	plan.ID = types.StringValue(updatedGroup.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting group",
			fmt.Sprintf("Could not delete group ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
