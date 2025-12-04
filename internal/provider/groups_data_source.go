// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/groups"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// GroupDataSource defines the data source implementation.
type GroupDataSource struct {
	client *groups.Client
}

// GroupDataSourceModel describes the data source data model.
type GroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	UserIDs     types.List   `tfsdk:"user_ids"`
	Permissions types.Object `tfsdk:"permissions"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (d *GroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Group data source for OpenWebUI",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group identifier",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the group to look up",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the group",
			},
			"user_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of user IDs in the group",
			},
			"permissions": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Permissions for the group",
				Attributes: map[string]schema.Attribute{
					"workspace": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"models":    schema.BoolAttribute{Computed: true},
							"knowledge": schema.BoolAttribute{Computed: true},
							"prompts":   schema.BoolAttribute{Computed: true},
							"tools":     schema.BoolAttribute{Computed: true},
						},
					},
					"chat": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"file_upload":         schema.BoolAttribute{Computed: true},
							"delete":              schema.BoolAttribute{Computed: true},
							"edit":                schema.BoolAttribute{Computed: true},
							"temporary":           schema.BoolAttribute{Computed: true},
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
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"public_models":    schema.BoolAttribute{Required: true},
							"public_knowledge": schema.BoolAttribute{Required: true},
							"public_prompts":   schema.BoolAttribute{Required: true},
							"public_tools":     schema.BoolAttribute{Required: true},
							"public_notes":     schema.BoolAttribute{Required: true},
						},
					},
					"features": schema.SingleNestedAttribute{
						Computed: true,
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
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the group was created",
			},
			"updated_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the group was last updated",
			},
		},
	}
}

func (d *GroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get groups from API
	groups, err := d.client.List()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read groups, got error: %s", err))
		return
	}

	// Find the group with matching name
	var found bool
	for _, group := range groups {
		if group.Name == data.Name.ValueString() {
			// Convert API response to model
			data.ID = types.StringValue(group.ID)
			data.Description = types.StringValue(group.Description)
			data.CreatedAt = types.Int64Value(group.CreatedAt)
			data.UpdatedAt = types.Int64Value(group.UpdatedAt)

			// Handle user IDs
			userIDs, diags := types.ListValueFrom(ctx, types.StringType, group.UserIDs)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.UserIDs = userIDs

			// Handle permissions
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

				featureAttrs := map[string]attr.Value{
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

				featureObj, diags := types.ObjectValue(
					map[string]attr.Type{
						"direct_tool_servers": types.BoolType,
						"web_search":          types.BoolType,
						"image_generation":    types.BoolType,
						"code_interpreter":    types.BoolType,
						"notes":               types.BoolType,
					},
					featureAttrs,
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
						"features":  types.ObjectType{AttrTypes: featureObj.AttributeTypes(ctx)},
					},
					map[string]attr.Value{
						"workspace": workspaceObj,
						"chat":      chatObj,
						"sharing":   sharingObj,
						"features":  featureObj,
					},
				)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				data.Permissions = permissionsObj
			}

			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Group Not Found",
			fmt.Sprintf("No group found with name: %s", data.Name.ValueString()),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
