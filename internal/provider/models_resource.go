// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/models"
)

var (
	_ resource.Resource                = &ModelResource{}
	_ resource.ResourceWithImportState = &ModelResource{}
)

func NewModelResource() resource.Resource {
	return &ModelResource{}
}

type ModelResource struct {
	client *models.Client
}

func (r *ModelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *ModelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	client, ok := clients["models"].(*models.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *models.Client, got: %T. Please report this issue to the provider developers.", clients["models"]),
		)
		return
	}

	r.client = client
}

func (r *ModelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a model in OpenWebUI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the model.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who created the model.",
				Computed:    true,
			},
			"base_model_id": schema.StringAttribute{
				Description: "The ID of the base model.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the model.",
				Required:    true,
			},
			"params": schema.SingleNestedAttribute{
				Description: "Model parameters.",
				Computed:    true,
				Optional:    true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(map[string]attr.Type{
					"system":            types.StringType,
					"stream_response":   types.BoolType,
					"temperature":       types.Float64Type,
					"reasoning_effort":  types.StringType,
					"top_p":             types.Float64Type,
					"top_k":             types.Int64Type,
					"min_p":             types.Float64Type,
					"max_tokens":        types.Int64Type,
					"seed":              types.Int64Type,
					"frequency_penalty": types.Int64Type,
					"repeat_last_n":     types.Int64Type,
					"num_ctx":           types.Int64Type,
					"num_batch":         types.Int64Type,
					"num_keep":          types.Int64Type,
					"function_calling":  types.StringType,
				}, map[string]attr.Value{
					"system":            types.StringNull(),
					"stream_response":   types.BoolNull(),
					"temperature":       types.Float64Null(),
					"reasoning_effort":  types.StringNull(),
					"top_p":             types.Float64Null(),
					"top_k":             types.Int64Null(),
					"min_p":             types.Float64Null(),
					"max_tokens":        types.Int64Null(),
					"seed":              types.Int64Null(),
					"frequency_penalty": types.Int64Null(),
					"repeat_last_n":     types.Int64Null(),
					"num_ctx":           types.Int64Null(),
					"num_batch":         types.Int64Null(),
					"num_keep":          types.Int64Null(),
					"function_calling":  types.StringNull(),
				})),
				Attributes: map[string]schema.Attribute{
					"system": schema.StringAttribute{
						Description: "System prompt for the model.",
						Optional:    true,
					},
					"stream_response": schema.BoolAttribute{
						Description: "Whether to stream responses.",
						Optional:    true,
					},
					"temperature": schema.Float64Attribute{
						Description: "Sampling temperature.",
						Optional:    true,
					},
					"reasoning_effort": schema.StringAttribute{
						Description: "Reasoning effort level. If set, must be one of: 'low', 'medium', 'high'.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("low", "medium", "high"),
						},
					},
					"top_p": schema.Float64Attribute{
						Description: "Top-p sampling parameter.",
						Optional:    true,
					},
					"top_k": schema.Int64Attribute{
						Description: "Top-k sampling parameter.",
						Optional:    true,
					},
					"min_p": schema.Float64Attribute{
						Description: "Minimum probability threshold.",
						Optional:    true,
					},
					"max_tokens": schema.Int64Attribute{
						Description: "Maximum number of tokens to generate.",
						Optional:    true,
					},
					"seed": schema.Int64Attribute{
						Description: "Random seed for reproducibility.",
						Optional:    true,
					},
					"frequency_penalty": schema.Int64Attribute{
						Description: "Frequency penalty.",
						Optional:    true,
					},
					"repeat_last_n": schema.Int64Attribute{
						Description: "Number of tokens to consider for repetition penalty.",
						Optional:    true,
					},
					"num_ctx": schema.Int64Attribute{
						Description: "Context window size.",
						Optional:    true,
					},
					"num_batch": schema.Int64Attribute{
						Description: "Batch size for processing.",
						Optional:    true,
					},
					"num_keep": schema.Int64Attribute{
						Description: "Number of tokens to keep from prompt.",
						Optional:    true,
					},
					"function_calling": schema.StringAttribute{
						Description: "Enables function calling support; set to 'native' for API native support, otherwise omit.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("native"),
						},
					},
				},
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Model metadata.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"profile_image_url": schema.StringAttribute{
						Description: "URL for the model's profile image.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("/static/favicon.png"),
					},
					"description": schema.StringAttribute{
						Description: "Description of the model.",
						Optional:    true,
					},
					"capabilities": schema.SingleNestedAttribute{
						Description: "Model capabilities.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"vision": schema.BoolAttribute{
								Description: "Whether the model supports vision tasks.",
								Optional:    true,
							},
							"usage": schema.BoolAttribute{
								Description: "Whether to track usage statistics.",
								Optional:    true,
							},
							"citations": schema.BoolAttribute{
								Description: "Whether the model supports citations.",
								Optional:    true,
							},
						},
					},
					"tags": schema.ListNestedAttribute{
						Description: "List of tags.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the tag.",
									Required:    true,
								},
							},
						},
					},
					"filter_ids": schema.SetAttribute{
						Description: "List of filter IDs.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"access_control": schema.SingleNestedAttribute{
				Description:   "Access control settings.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Object{AccessControlDefaultModifier{}},
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
			"is_active": schema.BoolAttribute{
				Description: "Whether the model is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.Int64Attribute{
				Description: "Timestamp when the model was created.",
				Computed:    true,
			},
			"updated_at": schema.Int64Attribute{
				Description: "Timestamp when the model was last updated.",
				Computed:    true,
			},
			"is_private": schema.BoolAttribute{
				Description:         "Whether the model is private.",
				MarkdownDescription: "Whether the model is private. `access_control` must be unset when this is set to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.client.CreateModel(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
	}

	// Ensure the ID is set in the state
	if model.ID.IsNull() {
		resp.Diagnostics.AddError("Error creating model", "Model ID is null after creation")
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.client.GetModel(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading model", err.Error())
		return
	}

	// Ensure the ID is preserved
	if model.ID.IsNull() {
		model.ID = state.ID
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.Model
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure we use the existing ID for the update
	plan.ID = state.ID

	model, err := r.client.UpdateModel(state.ID.ValueString(), &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating model", err.Error())
		return
	}

	// Ensure the ID is preserved
	if model.ID.IsNull() {
		model.ID = state.ID
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteModel(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting model", err.Error())
		return
	}
}

func (r *ModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// AccessControlDefaultModifier is a custom plan modifier for the access_control attribute.
type AccessControlDefaultModifier struct{}

func (m AccessControlDefaultModifier) Description(ctx context.Context) string {
	return "Ensures access_control is set to default values when is_private is true."
}

func (m AccessControlDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return "Ensures access_control is set to default values when is_private is true."
}

// PlanModifyObject implements the plan modification logic.
func (m AccessControlDefaultModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	var isPrivate types.Bool
	req.Plan.GetAttribute(ctx, path.Root("is_private"), &isPrivate)

	var accessControl types.Object
	req.Plan.GetAttribute(ctx, path.Root("access_control"), &accessControl)

	if !isPrivate.ValueBool() && !(accessControl.IsNull() || accessControl.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(path.Root("is_private"), "Invalid Value", "is_private must be true when access_control is specified.")
		resp.Diagnostics.AddAttributeError(path.Root("access_control"), "Invalid Value", "access_control cannot be specified when is_private is false.")
		return
	}

	attributeTypes := map[string]attr.Type{
		"read":  types.ObjectType{AttrTypes: map[string]attr.Type{"group_ids": types.ListType{ElemType: types.StringType}, "user_ids": types.ListType{ElemType: types.StringType}}},
		"write": types.ObjectType{AttrTypes: map[string]attr.Type{"group_ids": types.ListType{ElemType: types.StringType}, "user_ids": types.ListType{ElemType: types.StringType}}},
	}

	if isPrivate.ValueBool() && (accessControl.IsUnknown() || accessControl.IsNull()) {
		emptyList, listDiags := types.ListValueFrom(context.Background(), types.StringType, []string{})

		resp.Diagnostics.Append(listDiags...)

		readPermissions, readDiags := types.ObjectValue(map[string]attr.Type{"group_ids": types.ListType{ElemType: types.StringType}, "user_ids": types.ListType{ElemType: types.StringType}}, map[string]attr.Value{
			"group_ids": emptyList,
			"user_ids":  emptyList,
		})
		resp.Diagnostics.Append(readDiags...)
		writePermissions, writeDiags := types.ObjectValue(map[string]attr.Type{"group_ids": types.ListType{ElemType: types.StringType}, "user_ids": types.ListType{ElemType: types.StringType}}, map[string]attr.Value{
			"group_ids": emptyList,
			"user_ids":  emptyList,
		})
		resp.Diagnostics.Append(writeDiags...)

		defaultAccessControl, diags := types.ObjectValue(attributeTypes, map[string]attr.Value{
			"read":  readPermissions,
			"write": writePermissions,
		})
		resp.Diagnostics.Append(diags...)
		resp.PlanValue = defaultAccessControl
	}
	if !isPrivate.ValueBool() {
		resp.PlanValue = types.ObjectNull(attributeTypes)
	}
}
