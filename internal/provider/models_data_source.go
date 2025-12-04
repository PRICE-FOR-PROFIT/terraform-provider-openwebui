// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/models"
)

var (
	_ datasource.DataSource = &ModelDataSource{}
)

func NewModelDataSource() datasource.DataSource {
	return &ModelDataSource{}
}

type ModelDataSource struct {
	client *models.Client
}

func (d *ModelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *ModelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a model by name.",
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
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the model.",
				Computed:    true,
			},
			"params": schema.SingleNestedAttribute{
				Description: "Model parameters.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"system": schema.StringAttribute{
						Description: "System prompt for the model.",
						Computed:    true,
					},
					"stream_response": schema.BoolAttribute{
						Description: "Whether to stream responses.",
						Computed:    true,
					},
					"temperature": schema.Float64Attribute{
						Description: "Sampling temperature.",
						Computed:    true,
					},
					"reasoning_effort": schema.StringAttribute{
						Description: "Reasoning effort level.",
						Computed:    true,
					},
					"top_p": schema.Float64Attribute{
						Description: "Top-p sampling parameter.",
						Computed:    true,
					},
					"top_k": schema.Int64Attribute{
						Description: "Top-k sampling parameter.",
						Computed:    true,
					},
					"min_p": schema.Float64Attribute{
						Description: "Minimum probability threshold.",
						Computed:    true,
					},
					"max_tokens": schema.Int64Attribute{
						Description: "Maximum number of tokens to generate.",
						Computed:    true,
					},
					"seed": schema.Int64Attribute{
						Description: "Random seed for reproducibility.",
						Computed:    true,
					},
					"frequency_penalty": schema.Int64Attribute{
						Description: "Frequency penalty.",
						Computed:    true,
					},
					"repeat_last_n": schema.Int64Attribute{
						Description: "Number of tokens to consider for repetition penalty.",
						Computed:    true,
					},
					"num_ctx": schema.Int64Attribute{
						Description: "Context window size.",
						Computed:    true,
					},
					"num_batch": schema.Int64Attribute{
						Description: "Batch size for processing.",
						Computed:    true,
					},
					"num_keep": schema.Int64Attribute{
						Description: "Number of tokens to keep from prompt.",
						Computed:    true,
					},
					"function_calling": schema.StringAttribute{
						Description: "Type of function calling support (set to 'native' if enabled).",
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("native"),
						},
					},
				},
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Model metadata.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"profile_image_url": schema.StringAttribute{
						Description: "URL for the model's profile image.",
						Computed:    true,
					},
					"description": schema.StringAttribute{
						Description: "Description of the model.",
						Computed:    true,
					},
					"capabilities": schema.SingleNestedAttribute{
						Description: "Model capabilities.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"vision": schema.BoolAttribute{
								Description: "Whether the model supports vision tasks.",
								Computed:    true,
							},
							"usage": schema.BoolAttribute{
								Description: "Whether to track usage statistics.",
								Computed:    true,
							},
							"citations": schema.BoolAttribute{
								Description: "Whether the model supports citations.",
								Computed:    true,
							},
						},
					},
					"tags": schema.ListNestedAttribute{
						Description: "List of tags.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the tag.",
									Computed:    true,
								},
							},
						},
					},
					"filter_ids": schema.ListAttribute{
						Description: "List of filter IDs.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"access_control": schema.SingleNestedAttribute{
				Description: "Access control settings.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"read": schema.SingleNestedAttribute{
						Description: "Read access settings.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"group_ids": schema.ListAttribute{
								Description: "List of group IDs with read access.",
								Computed:    true,
								ElementType: types.StringType,
							},
							"user_ids": schema.ListAttribute{
								Description: "List of user IDs with read access.",
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
					"write": schema.SingleNestedAttribute{
						Description: "Write access settings.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"group_ids": schema.ListAttribute{
								Description: "List of group IDs with write access.",
								Computed:    true,
								ElementType: types.StringType,
							},
							"user_ids": schema.ListAttribute{
								Description: "List of user IDs with write access.",
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the model is active.",
				Computed:    true,
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
				Description: "Whether the model is private.",
				Computed:    true,
			},
		},
	}
}

func (d *ModelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected map[string]interface{}, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	client, ok := clients["models"].(*models.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *models.Client, got: %T. Please report this issue to the provider developers.", clients["models"]),
		)
		return
	}

	d.client = client
}

func (d *ModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.Model
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get specific model
	foundModel, err := d.client.GetModel(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading model", err.Error())
		return
	}
	diags = resp.State.Set(ctx, foundModel)
	resp.Diagnostics.Append(diags...)

	if foundModel == nil {
		resp.Diagnostics.AddError(
			"Error reading model",
			fmt.Sprintf("No model found with name: %s", config.ID.ValueString()),
		)
		return
	}
}
