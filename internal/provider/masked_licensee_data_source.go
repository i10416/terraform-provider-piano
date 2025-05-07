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
	_ datasource.DataSource              = &LicenseeDataSource{}
	_ datasource.DataSourceWithConfigure = &LicenseeDataSource{}
)

func NewMaskedLicenseeDataSource() datasource.DataSource {
	return &MaskedLicenseeDataSource{}
}

// MaskedLicenseeDataSource defines the data source implementation.
type MaskedLicenseeDataSource struct {
	client *piano_publisher.Client
}

// MaskedLicenseeDataSourceModel describes the data source data model.
type MaskedLicenseeDataSourceModel struct {
	// required
	Aid        types.String `tfsdk:"aid"`
	Name       types.String `tfsdk:"name"`
	LicenseeId types.String `tfsdk:"licensee_id"`
	// optional
	Description types.String             `tfsdk:"description"`
	LogoUrl     types.String             `tfsdk:"logo_url"`
	Managers    []ManagerDataSourceModel `tfsdk:"managers"`
}

func (d *MaskedLicenseeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_masked_licensee"
}

func (d *MaskedLicenseeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MaskedLicensee data source. This data source is used to get the masked licensee details.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				MarkdownDescription: "piano application id",
				Required:            true,
			},
			"licensee_id": schema.StringAttribute{
				MarkdownDescription: "The public ID of the licensee",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the licensee",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the licensee",
				Computed:            true,
			},
			"logo_url": schema.StringAttribute{
				MarkdownDescription: "A relative URL of the licensee's logo",
				Computed:            true,
			},
			"managers": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uid": schema.StringAttribute{
							Required: true,
						},
						"first_name": schema.StringAttribute{
							Required: true,
						},
						"last_name": schema.StringAttribute{
							Required: true,
						},
						"personal_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (d *MaskedLicenseeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *MaskedLicenseeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MaskedLicenseeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.client.GetPublisherLicensingLicenseeGet(ctx, &piano_publisher.GetPublisherLicensingLicenseeGetParams{
		Aid:        data.Aid.ValueString(),
		LicenseeId: data.LicenseeId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch licensee, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.LicenseeResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data.Name = types.StringValue(result.Licensee.Name)
	data.Description = types.StringPointerValue(result.Licensee.Description)
	data.LogoUrl = types.StringPointerValue(result.Licensee.LogoUrl)
	managers := ManagerDataSourceFromData(result.Licensee.Managers)
	data.Managers = managers
	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
