// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &LicenseeDataSource{}
	_ datasource.DataSourceWithConfigure = &LicenseeDataSource{}
)

func NewLicenseeDataSource() datasource.DataSource {
	return &LicenseeDataSource{}
}

// LicenseeDataSource defines the data source implementation.
type LicenseeDataSource struct {
	client *piano_publisher.Client
}

// LicenseeDataSourceModel describes the data source data model.
type LicenseeDataSourceModel struct {
	// required
	Aid        types.String `tfsdk:"aid"`
	Name       types.String `tfsdk:"name"`
	LicenseeId types.String `tfsdk:"licensee_id"`
	// optional
	Description     types.String                    `tfsdk:"description"`
	LogoUrl         types.String                    `tfsdk:"logo_url"`
	Representatives []RepresentativeDataSourceModel `tfsdk:"representatives"`
	Managers        []ManagerDataSourceModel        `tfsdk:"managers"`
}

type ManagerDataSourceModel struct {
	UID          types.String `tfsdk:"uid"`           // The user's ID
	FirstName    types.String `tfsdk:"first_name"`    // The user's first name
	LastName     types.String `tfsdk:"last_name"`     // The user's last name
	PersonalName types.String `tfsdk:"personal_name"` // The user's personal name. Name and surname ordered as per locale
}

type RepresentativeDataSourceModel struct {
	Email types.String `tfsdk:"email"`
}

func (d *LicenseeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_licensee"
}

func (d *LicenseeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Licensee data source. Licensee is a company that has access to resources in the app.",
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
			"representatives": schema.ListAttribute{
				Computed: true,
				ElementType: basetypes.ObjectType{
					AttrTypes: map[string]attr.Type{
						"email": basetypes.StringType{},
					},
				},
			},
			"managers": schema.ListNestedAttribute{
				Computed: true,
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

func (d *LicenseeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PianoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = &client.publisherClient
}

func (d *LicenseeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LicenseeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.client.GetPublisherLicensingLicenseeGet(ctx, &piano_publisher.GetPublisherLicensingLicenseeGetParams{
		Aid:        state.Aid.ValueString(),
		LicenseeId: state.LicenseeId.ValueString(),
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

	state.Name = types.StringValue(result.Licensee.Name)
	state.Description = types.StringPointerValue(result.Licensee.Description)
	state.LogoUrl = types.StringPointerValue(result.Licensee.LogoUrl)
	representatives := []RepresentativeDataSourceModel{}
	for _, r := range result.Licensee.Representatives {
		rep := RepresentativeDataSourceModel{}
		rep.Email = types.StringValue(r.Email)
		representatives = append(representatives, rep)
	}
	state.Representatives = representatives
	managers := ManagerDataSourceFromData(result.Licensee.Managers)
	state.Managers = managers
	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func ManagerDataSourceFromData(items []piano_publisher.LicenseeManager) []ManagerDataSourceModel {
	managers := []ManagerDataSourceModel{}
	for _, item := range items {
		manager := ManagerDataSourceModel{}
		manager.FirstName = types.StringValue(item.FirstName)
		manager.LastName = types.StringValue(item.LastName)
		manager.PersonalName = types.StringValue(item.PersonalName)
		manager.UID = types.StringValue(item.Uid)
		managers = append(managers, manager)
	}
	return managers
}
