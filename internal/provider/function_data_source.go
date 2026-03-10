// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-openwebui/internal/provider/client/functions"
)

var (
	_ datasource.DataSource = &FunctionDataSource{}
)

func NewFunctionDataSource() datasource.DataSource {
	return &FunctionDataSource{}
}

type FunctionDataSource struct {
	client *functions.Client
}

func (d *FunctionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (d *FunctionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a function by ID.",
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
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the function.",
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/code of the function.",
				Computed:    true,
			},
			"meta": schema.SingleNestedAttribute{
				Description: "Function metadata.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "Description of the function.",
						Computed:    true,
					},
					"manifest": schema.MapAttribute{
						Description: "Function manifest metadata.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the function is active.",
				Computed:    true,
			},
			"is_global": schema.BoolAttribute{
				Description: "Whether the function is global.",
				Computed:    true,
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

func (d *FunctionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	client, ok := clients["functions"].(*functions.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *functions.Client, got: %T. Please report this issue to the provider developers.", clients["functions"]),
		)
		return
	}

	d.client = client
}

func (d *FunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config functions.Function
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get specific function
	foundFunction, err := d.client.Get(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading function", err.Error())
		return
	}

	if foundFunction == nil {
		resp.Diagnostics.AddError(
			"Error reading function",
			fmt.Sprintf("No function found with ID: %s", config.ID.ValueString()),
		)
		return
	}

	// Convert API response to Terraform model
	state := functions.APIToFunction(foundFunction)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
