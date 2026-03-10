// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/tools"
)

var (
	_ datasource.DataSource = &ToolDataSource{}
)

func NewToolDataSource() datasource.DataSource {
	return &ToolDataSource{}
}

type ToolDataSource struct {
	client *tools.Client
}

func (d *ToolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (d *ToolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a tool by ID.",
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
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/code of the tool.",
				Computed:    true,
			},
			"specs": schema.StringAttribute{
				Description: "Tool specifications as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Tool metadata.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "Description of the tool.",
						Computed:    true,
					},
					"manifest": schema.MapAttribute{
						Description: "Tool manifest metadata.",
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

func (d *ToolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	client, ok := clients["tools"].(*tools.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tools.Client, got: %T. Please report this issue to the provider developers.", clients["tools"]),
		)
		return
	}

	d.client = client
}

func (d *ToolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tools.Tool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get specific tool
	foundTool, err := d.client.Get(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading tool", err.Error())
		return
	}

	if foundTool == nil {
		resp.Diagnostics.AddError(
			"Error reading tool",
			fmt.Sprintf("No tool found with ID: %s", config.ID.ValueString()),
		)
		return
	}

	// Convert API response to Terraform model
	state := tools.APIToTool(foundTool)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
