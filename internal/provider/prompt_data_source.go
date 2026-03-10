// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/prompts"
)

var (
	_ datasource.DataSource = &PromptDataSource{}
)

func NewPromptDataSource() datasource.DataSource {
	return &PromptDataSource{}
}

type PromptDataSource struct {
	client *prompts.Client
}

func (d *PromptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (d *PromptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a prompt by command/ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The command/ID of the prompt (e.g., '/summarize').",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who created the prompt.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the prompt.",
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/template of the prompt.",
				Computed:    true,
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
			"timestamp": schema.Int64Attribute{
				Description: "Timestamp when the prompt was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *PromptDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	client, ok := clients["prompts"].(*prompts.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *prompts.Client, got: %T. Please report this issue to the provider developers.", clients["prompts"]),
		)
		return
	}

	d.client = client
}

func (d *PromptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config prompts.Prompt
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get specific prompt using id as command
	foundPrompt, err := d.client.Get(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading prompt", err.Error())
		return
	}

	if foundPrompt == nil {
		resp.Diagnostics.AddError(
			"Error reading prompt",
			fmt.Sprintf("No prompt found with command: %s", config.ID.ValueString()),
		)
		return
	}

	// Convert API response to Terraform model
	state := prompts.APIToPrompt(foundPrompt)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
