// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &AppDataSource{}
	_ datasource.DataSourceWithConfigure = &AppDataSource{}
)

func NewAppDataSource() datasource.DataSource {
	return &AppDataSource{}
}

// LicenseeDataSource defines the data source implementation.
type AppDataSource struct {
	client *piano_publisher.Client
}

// AppDataSourceModel describes the data source data model.

type AppDataSourceModel struct {
	Aid          types.String `tfsdk:"aid"`           // The application ID
	DefaultLang  types.String `tfsdk:"default_lang"`  // The default language
	EmailLang    types.String `tfsdk:"email_lang"`    // The email language
	Details      types.String `tfsdk:"details"`       // The application details
	Email        types.String `tfsdk:"email"`         // Email address associated with this app
	Name         types.String `tfsdk:"name"`          // The application name
	UserProvider types.String `tfsdk:"user_provider"` // The user token provider
	URL          types.String `tfsdk:"url"`           // The application website
	Logo1        types.String `tfsdk:"logo1"`         // Primary image displayed within the dashboard
	Logo2        types.String `tfsdk:"logo2"`         // Secondary image displayed within the ticket
	State        types.String `tfsdk:"state"`         // Current state of the app
}

func (*AppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (*AppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "piano app source",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				MarkdownDescription: "piano application id",
				Required:            true,
			},
			"default_lang": schema.StringAttribute{
				MarkdownDescription: "default language",
				Computed:            true,
			},
			"email_lang": schema.StringAttribute{
				MarkdownDescription: "email language",
				Computed:            true,
			},
			"details": schema.StringAttribute{
				MarkdownDescription: "application details",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "email address associated with this app",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "application name",
				Computed:            true,
			},
			"user_provider": schema.StringAttribute{
				MarkdownDescription: "user token provider",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "application website",
				Computed:            true,
			},
			"logo1": schema.StringAttribute{
				MarkdownDescription: "primary image displayed within the dashboard",
				Computed:            true,
			},
			"logo2": schema.StringAttribute{
				MarkdownDescription: "secondary image displayed within the ticket",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "current state of the app",
				Computed:            true,
			},
		},
	}
}

func (d *AppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AppDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	response, err := d.client.GetPublisherAppGet(ctx, &piano_publisher.GetPublisherAppGetParams{
		Aid: state.Aid.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch licensee, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.AppResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)

	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	state.Name = types.StringValue(result.App.Name)
	state.Aid = types.StringValue(result.App.Aid)
	state.DefaultLang = types.StringValue(result.App.DefaultLang)
	state.URL = types.StringValue(result.App.Url)
	state.Email = types.StringValue(result.App.Email)
	state.EmailLang = types.StringValue(result.App.EmailLang)
	state.Details = types.StringPointerValue(result.App.Details)
	state.Logo1 = types.StringValue(result.App.Logo1)
	state.Logo2 = types.StringPointerValue(result.App.Logo2)
	state.State = types.StringValue(string(result.App.State))
	state.UserProvider = types.StringValue(string(result.App.UserProvider))
	tflog.Trace(ctx, "read an app data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
